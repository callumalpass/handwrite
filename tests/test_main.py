import unittest
import os
import tempfile
import shutil
from unittest.mock import patch, MagicMock
from PIL import Image
import yaml

from handwriting_processor.main import (
    get_config_path,
    load_config,
    get_images,
    render_markdown,
    setup_config
)


class TestHandwritingProcessor(unittest.TestCase):
    
    def setUp(self):
        self.test_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.test_dir)
    
    def test_get_config_path_with_custom_path(self):
        custom_path = "/tmp/custom_config.yaml"
        with patch('os.path.exists', return_value=True):
            result = get_config_path(custom_path)
            self.assertEqual(result, custom_path)
    
    def test_get_config_path_default(self):
        result = get_config_path()
        expected = os.path.expanduser("~/.config/handwrite/config.yaml")
        self.assertEqual(result, expected)
    
    def test_load_config_file_exists(self):
        config_file = os.path.join(self.test_dir, "config.yaml")
        test_config = {
            "gemini": {
                "model": "test-model",
                "prompt": "test prompt"
            },
            "template": {
                "path": "/test/path"
            }
        }
        
        with open(config_file, 'w') as f:
            yaml.dump(test_config, f)
        
        result = load_config(config_file)
        self.assertEqual(result["gemini"]["model"], "test-model")
        self.assertEqual(result["gemini"]["prompt"], "test prompt")
    
    def test_load_config_file_not_exists(self):
        non_existent_path = "/non/existent/path.yaml"
        result = load_config(non_existent_path)
        
        self.assertIn("gemini", result)
        self.assertIn("template", result)
        self.assertEqual(result["gemini"]["model"], "gemini-1.5-pro")
    
    def test_get_images_unsupported_format(self):
        test_file = os.path.join(self.test_dir, "test.txt")
        with open(test_file, 'w') as f:
            f.write("test content")
        
        result = get_images(test_file)
        self.assertEqual(result, [])
    
    @patch('handwriting_processor.main.fitz.open')
    def test_get_images_pdf_error(self, mock_fitz_open):
        mock_fitz_open.side_effect = Exception("PDF error")
        
        result = get_images("/fake/path.pdf")
        self.assertEqual(result, [])
    
    def test_render_markdown_basic(self):
        template_dir = os.path.join(self.test_dir, "templates")
        os.makedirs(template_dir)
        
        template_file = os.path.join(template_dir, "test_template.md")
        with open(template_file, 'w') as f:
            f.write("# {{ title }}\n\n{{ content }}")
        
        output_file = os.path.join(self.test_dir, "output.md")
        
        render_markdown(template_file, output_file, "utf-8", title="Test Title", content="Test content")
        
        with open(output_file, 'r') as f:
            result = f.read()
        
        self.assertIn("# Test Title", result)
        self.assertIn("Test content", result)
    
    def test_setup_config_creates_directory(self):
        config_dir = os.path.join(self.test_dir, ".config", "handwrite")
        config_path = os.path.join(config_dir, "config.yaml")
        
        with patch('handwriting_processor.main.get_config_path', return_value=config_path):
            setup_config()
        
        self.assertTrue(os.path.exists(config_path))
        
        with open(config_path, 'r') as f:
            config = yaml.safe_load(f)
        
        self.assertIn("gemini", config)
        self.assertIn("template", config)
    
    def test_setup_config_already_exists(self):
        config_path = os.path.join(self.test_dir, "existing_config.yaml")
        with open(config_path, 'w') as f:
            f.write("existing: config")
        
        with patch('handwriting_processor.main.get_config_path', return_value=config_path):
            setup_config()
        
        # Config should remain unchanged
        with open(config_path, 'r') as f:
            content = f.read()
        
        self.assertEqual(content, "existing: config")


if __name__ == '__main__':
    unittest.main()