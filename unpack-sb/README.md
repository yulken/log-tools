# Unpack SB

A specialized automation script to extract logs from a specific nested structure typically used in system bundles.

## How it works

1.  Looks for a folder named `nodes` in the current directory.
2.  Iterates through all `.zip` files.
3.  Extracts each ZIP.
4.  If a `scc` subdirectory exists inside the extracted folder, it automatically extracts any `.txz` files found within it.
5.  Moves the processed folders to a root directory called `unpacked_nodes`.

## Usage

Simply run the binary inside the directory containing the `nodes` folder:
```bash
./unpack-sb
```
