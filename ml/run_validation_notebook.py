from pathlib import Path

import nbformat
from nbclient import NotebookClient


root = Path(__file__).resolve().parents[1]
path = root / "ml/validation_1000.ipynb"
notebook = nbformat.read(path, as_version=4)
NotebookClient(notebook, timeout=600, kernel_name="python3").execute(cwd=root)
print(f"executed {path} successfully")
