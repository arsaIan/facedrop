# Wip 
### Setup
```
git clone
cd facedrop
git submodule init
git submodule update
```
### Deepface setup using conda
Requires python < 3.11
```
conda create conda create --name venv python=3.10.10

conda activate venv

pip install -e .

cd scripts

./service.sh



