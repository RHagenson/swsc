# swsc

`swsc` implements the Sliding-Window Site Characteristics (SWSC) method as described in <https://doi.org/10.1093/molbev/msy069>

Initial write was based on [PFinderUCE-SWSC-EN]

[PFinderUCE-SWSC-EN]: https://github.com/Tagliacollo/PFinderUCE-SWSC-EN

## Usage

### Installation

1. Install the Go language following instructions at: https://golang.org/
2. Run `go get -u bitbucket.org/rhagenson/swsc/...`
3. Run `swsc` by either:
    + Calling it directly if you added`$GOPATH/bin/` to your `$PATH`
    + Navigating to `$GOPATH/src/bitbucket.org/rhagenson/swsc/` and running `go build main.go && ./swsc`

### Running

Both `input`,`output`, and at least one metric (`--gc` or `--entropy`) must be set. See `swsc --help` for details.

### Reporting Errors

If you have found an error, or this tools does not work for you, please create an issue at: https://bitbucket.org/RHagenson/swsc/issues with details on when the error occurred, what the error states, and what was expected to occur, if known.

## Input

`swsc` reads a single nexus file processing two blocks:

1. `DATA`, containing the UCE markers (unique by ID)
2. `SETS`, containing the UCE locations (unique by ID, with inclusive range)

Example (any `...` denotes truncated content, see [PFinderUCE-SWSC-EN] for full file):

```text
#NEXUS

BEGIN DATA;
DIMENSIONS  NTAX=10 NCHAR=5786;
FORMAT DATATYPE=DNA GAP=- MISSING=?;
MATRIX

sp1    AGAAAC...TGCAAAG
...
;

END;

BEGIN SETS;

    [loci]
    CHARSET chr_2828 = 1-376;
    CHARSET chr_4312 = 377-627;
    ...

    CHARPARTITION loci = 1:chr_2828, 2:chr_4312...;

END;
```

## Output

`swsc` writes a single `.csv` file containing the chosen characteristics for each site of the UCEs. It can also produce a `.cfg` for use by PartitionFinder2 by using the appropriate flag.
