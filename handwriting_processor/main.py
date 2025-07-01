import argparse
import os
import yaml
import logging
import sys
import glob
from datetime import datetime
from dotenv import load_dotenv
import google.generativeai as genai
import fitz  # PyMuPDF
from PIL import Image
from jinja2 import Environment, FileSystemLoader
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

def get_config_path(config_arg=None):
    """Determines the path to the configuration file."""
    if config_arg and os.path.exists(config_arg):
        return config_arg
    
    default_config_dir = os.path.expanduser("~/.config/handwrite")
    default_config_path = os.path.join(default_config_dir, "config.yaml")
    
    return default_config_path

def setup_config():
    """Creates the default configuration file if it doesn't exist."""
    config_path = get_config_path()
    if os.path.exists(config_path):
        logger.info(f"Config file already exists at {config_path}")
        return

    os.makedirs(os.path.dirname(config_path), exist_ok=True)
    default_config = {
        "gemini": {
            "model": "gemini-1.5-pro",
            "prompt": (
                "Extract the handwritten text from this image.\n"
                "- Use $ for LaTeX, not ```latex.\n"
                "- Transcribe the text exactly as it appears.\n"
                "- The output must be only the transcribed Markdown, with no additional commentary."
            )
        },
        "template": {
            "path": os.path.join(os.path.dirname(__file__), "templates/note_template.md"),
            "variables": {}
        },
        "output": {
            "format": "markdown",
            "encoding": "utf-8"
        }
    }
    with open(config_path, 'w') as f:
        yaml.dump(default_config, f, default_flow_style=False)
    logger.info(f"Created default config at {config_path}")

def load_config(config_path):
    """Loads the configuration from a YAML file."""
    if not os.path.exists(config_path):
        logger.warning(f"Config file not found at {config_path}. Using default settings.")
        return {
            "gemini": {
                "model": "gemini-1.5-pro",
                "prompt": "Extract the handwritten text from this image."
            },
            "template": {
                "path": os.path.join(os.path.dirname(__file__), "templates/note_template.md"),
                "variables": {}
            },
            "output": {
                "format": "markdown",
                "encoding": "utf-8"
            }
        }
    with open(config_path, 'r') as f:
        return yaml.safe_load(f)

def configure_genai():
    """Configures the Generative AI model."""
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        logger.error("GEMINI_API_KEY not found in environment variables or .env file")
        raise ValueError("GEMINI_API_KEY not found in environment variables or .env file")
    try:
        genai.configure(api_key=api_key)
        logger.info("Successfully configured Gemini API")
    except Exception as e:
        logger.error(f"Failed to configure Gemini API: {e}")
        raise

def get_images(input_path):
    """Extracts images from a PDF or loads an image file."""
    file_extension = os.path.splitext(input_path)[1].lower()
    if file_extension == ".pdf":
        try:
            doc = fitz.open(input_path)
            images = []
            for page_num in range(len(doc)):
                page = doc.load_page(page_num)
                pix = page.get_pixmap()
                img = Image.frombytes("RGB", [pix.width, pix.height], pix.samples)
                images.append(img)
            return images
        except Exception as e:
            logger.error(f"Error processing PDF: {e}")
            return []
    elif file_extension in [".png", ".jpg", ".jpeg"]:
        try:
            return [Image.open(input_path)]
        except Exception as e:
            logger.error(f"Error processing image: {e}")
            return []
    else:
        logger.error(f"Unsupported file type: {file_extension}")
        return []

def ocr_image_with_gemini(image, model_name, prompt):
    """Performs OCR on a single image using Gemini."""
    try:
        model = genai.GenerativeModel(model_name)
        response = model.generate_content([prompt, image])
        return response.text
    except Exception as e:
        logger.error(f"Error during OCR: {e}")
        return f"Error: {e}"

def render_markdown(template_path, output_path, encoding="utf-8", **context):
    """Renders the extracted text into a Markdown template."""
    try:
        template_dir = os.path.dirname(template_path)
        template_name = os.path.basename(template_path)
        env = Environment(loader=FileSystemLoader(template_dir))
        template = env.get_template(template_name)
        content = template.render(**context)
        with open(output_path, 'w', encoding=encoding) as f:
            f.write(content)
        logger.info(f"Successfully created note at {output_path}")
    except Exception as e:
        logger.error(f"Error rendering Markdown: {e}")

def process_single_file(input_path, output_dir, config):
    """Process a single PDF or image file."""
    try:
        images = get_images(input_path)
        if not images:
            logger.error(f"No images could be extracted from {input_path}")
            return False
        
        logger.info(f"Extracted {len(images)} images from {input_path}")

        model_name = config["gemini"]["model"]
        prompt = config["gemini"]["prompt"]
        
        page_results = ["" for _ in images]
        with ThreadPoolExecutor() as executor:
            future_to_page = {executor.submit(ocr_image_with_gemini, image, model_name, prompt): i for i, image in enumerate(images)}
            
            with tqdm(total=len(images), desc=f"Processing {os.path.basename(input_path)}") as pbar:
                for future in as_completed(future_to_page):
                    page_index = future_to_page[future]
                    try:
                        page_text = future.result()
                        page_results[page_index] = page_text
                        logger.debug(f"Successfully processed page {page_index + 1}")
                    except Exception as exc:
                        error_msg = f"Error processing page {page_index + 1}: {exc}"
                        page_results[page_index] = error_msg
                        logger.error(error_msg)
                    pbar.update(1)

        full_text = "\n\n".join(page_results)

        if full_text.strip():
            input_filename = os.path.basename(input_path)
            absolute_input_path = os.path.abspath(input_path)
            relative_input_path = os.path.relpath(absolute_input_path, output_dir)
            processing_datetime = datetime.now().isoformat()
            page_count = len(images)

            output_filename = os.path.splitext(input_filename)[0] + ".md"
            output_path = os.path.join(output_dir, output_filename)

            context = {
                "content": full_text,
                "filename": input_filename,
                "absolute_pdf_path": absolute_input_path,
                "relative_pdf_path": relative_input_path,
                "source_path_absolute": absolute_input_path,
                "source_path_relative": relative_input_path,
                "datetime_processed": processing_datetime,
                "page_count": page_count,
                "model_used": model_name,
            }
            
            # Add custom template variables from config
            if "variables" in config["template"]:
                context.update(config["template"]["variables"])

            template_path = config["template"]["path"]
            encoding = config.get("output", {}).get("encoding", "utf-8")
            render_markdown(template_path, output_path, encoding, **context)
            return True
        else:
            logger.warning(f"No text extracted from {input_path}")
            return False
    except Exception as e:
        logger.error(f"Error processing {input_path}: {e}")
        return False

def process_command(args):
    """Handles the 'process' command."""
    # Handle both single files and batch processing
    input_paths = []
    
    if os.path.isfile(args.input_path):
        input_paths = [args.input_path]
    elif os.path.isdir(args.input_path):
        # Batch processing - find all PDF and image files
        for ext in ['*.pdf', '*.PDF', '*.png', '*.PNG', '*.jpg', '*.JPG', '*.jpeg', '*.JPEG']:
            input_paths.extend(glob.glob(os.path.join(args.input_path, ext)))
            input_paths.extend(glob.glob(os.path.join(args.input_path, '**', ext), recursive=True))
        
        if not input_paths:
            logger.error(f"No supported files found in directory {args.input_path}")
            sys.exit(1)
    else:
        logger.error(f"Input path not found: {args.input_path}")
        sys.exit(1)

    if not os.path.isdir(args.output_dir):
        logger.error(f"Output directory not found at {args.output_dir}")
        sys.exit(1)

    try:
        config_path = get_config_path(args.config)
        config = load_config(config_path)
        configure_genai()

        logger.info(f"Processing {len(input_paths)} file(s)")
        
        successful = 0
        failed = 0
        
        for input_path in input_paths:
            logger.info(f"Processing: {input_path}")
            if process_single_file(input_path, args.output_dir, config):
                successful += 1
            else:
                failed += 1
        
        logger.info(f"Processing complete: {successful} successful, {failed} failed")
        
        if failed > 0:
            sys.exit(1)

    except Exception as e:
        logger.error(f"An unexpected error occurred: {e}")
        sys.exit(1)

def config_command(args):
    """Handles the 'config' command."""
    if args.setup:
        setup_config()

def main():
    """Main function to orchestrate the workflow."""
    parser = argparse.ArgumentParser(description="A tool for processing handwritten notes.")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # Process command
    process_parser = subparsers.add_parser("process", help="Process a PDF or image file.")
    process_parser.add_argument("input_path", help="The path to the PDF/image file or directory containing files to process.")
    process_parser.add_argument("output_dir", help="The directory to save the output Markdown file.")
    process_parser.add_argument("--config", help="Path to a custom configuration file.")
    process_parser.set_defaults(func=process_command)

    # Config command
    config_parser = subparsers.add_parser("config", help="Manage the configuration file.")
    config_parser.add_argument("--setup", action="store_true", help="Create the default configuration file.")
    config_parser.set_defaults(func=config_command)

    args = parser.parse_args()
    args.func(args)

if __name__ == "__main__":
    main()
