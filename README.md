# Pretty Slurm - a slurm tools collection

**Pretty Slurm** is a collection of executables that aim to both *prettify existing* slurm commands and *add new* useful commands for everyday slurm use.

## Installation

### Linux x86

Download the pre-built binaries from the releases page. ex:
```
curl https://github.com/pierreadorni/pretty-slurm/releases/download/v0.1/pretty-slurm.zip --output pretty-slurm.zip
```

unzip the binaries into your preferred bin folder (example here with `~/.local/bin`)
```
unzip pretty-slurm.zip -d ~/.local/bin/
```

### Other Unix

We do not build *yet* the binaries for other platforms, but because the code is written in *go* it is quite easy to build yourself. Below is a minimal step-by-step guide.

1. If needed, install go following [the documentation](https://go.dev/doc/install).
2. clone the repo `git clone git@github.com:pierreadorni/pretty-slurm.git`
3. build the executables `cd pretty-slurm && go build -o dist ./cmd/...`
4. copy the binaries into your preferred bin folder `cp dist/* ~/.local/bin/`


## Available Commands

### `psload` 

Displays the nodes according to their GPU availability, highlights the node with the most combined VRAM available.

Options:
 - `-best`: show only the best node 
<!--


-->

<p align="center">
    <img src="https://github.com/user-attachments/assets/fa30195e-1029-4d17-bd2e-aa3abc5823cf" width="600"/>
</p>


### `psblame`

Displays the cluster usage statistics by user over the last 7 days. Statistics are:
 - Compute time (Hours)
 - CPU time (CPU.Hours)
 - GPU time (GPU.Hours)
 - VRAM time (Gb.Hours)

Arguments: 
 - `-days int` the number of days to look back (default 7)

<p align="center">
    <img src="https://github.com/user-attachments/assets/57460d25-8128-4285-b885-f3f7ca6174b3" width="600"/>
</p>