module github.com/pierreadorni/pretty-slurm

go 1.23.2

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/fatih/color v1.18.0
	github.com/mattn/go-runewidth v0.0.16
	github.com/rodaine/table v1.3.0
	internal/slurmapi v1.0.0
)

replace internal/slurmapi => ./internal/slurmapi

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
