"""
Copyright (c) 2023-2025 RapidaAI
Author: Prashant Srivastav <prashant@rapida.ai>

Licensed under GPL-2.0 with Rapida Additional Terms.
See LICENSE.md for details or contact sales@rapida.ai for commercial use.
"""
import logging
import os
import shutil
from collections.abc import Generator

from app.configs.storage_config import AssetStoreConfig
from app.storage.file_storage.base_storage import BaseStorage
_log = logging.getLogger("app.storage.file_storage.local_storage")

class LocalStorage(BaseStorage):
    """Implementation for local storage.
    """

    def __init__(self, config: AssetStoreConfig):
        super().__init__()
        folder = config.storage_path_prefix
        self.folder = folder

    def save(self, filename, data):
        if not self.folder or self.folder.endswith('/'):
            filename = self.folder + filename
        else:
            filename = self.folder + '/' + filename

        filename = os.path.expanduser(filename)
        folder = os.path.dirname(filename)
        os.makedirs(folder, exist_ok=True)

        with open(os.path.join(os.getcwd(), filename), "wb") as f:
            f.write(data)

    def load_once(self, filename: str) -> bytes:
        if not self.folder or self.folder.endswith('/'):
            filename = self.folder + filename
        else:
            filename = self.folder + '/' + filename
        
        filename = os.path.expanduser(filename)
        if not os.path.exists(filename):
            raise FileNotFoundError("File not found")

        with open(filename, "rb") as f:
            data = f.read()

        return data

    def load_stream(self, filename: str) -> Generator:
        def generate(filename: str = filename) -> Generator:
            if not self.folder or self.folder.endswith('/'):
                filename = self.folder + filename
            else:
                filename = self.folder + '/' + filename
            
            filename = os.path.expanduser(filename)
            if not os.path.exists(filename):
                raise FileNotFoundError("File not found")

            with open(filename, "rb") as f:
                while chunk := f.read(4096):  # Read in chunks of 4KB
                    yield chunk

        return generate()

    def download(self, filename, target_filepath):
        if not self.folder or self.folder.endswith('/'):
            filename = self.folder + filename
        else:
            filename = self.folder + '/' + filename
        
        filename = os.path.expanduser(filename)

        if not os.path.exists(filename):
            raise FileNotFoundError("File not found")

        shutil.copyfile(filename, target_filepath)

    def exists(self, filename):
        if not self.folder or self.folder.endswith('/'):
            filename = self.folder + filename
        else:
            filename = self.folder + '/' + filename
        filename = os.path.expanduser(filename)
        return os.path.exists(filename)

    def delete(self, filename):
        if not self.folder or self.folder.endswith('/'):
            filename = self.folder + filename
        else:
            filename = self.folder + '/' + filename
        filename = os.path.expanduser(filename)
        if os.path.exists(filename):
            os.remove(filename)
