# swsc

`swsc` implements the Sliding-Window Site Characteristics (SWSC) method as described in <https://doi.org/10.1093/molbev/msy069>

Sample data taken from: [PFinderUCE-SWSC-EN]

[PFinderUCE-SWSC-EN]: https://github.com/Tagliacollo/PFinderUCE-SWSC-EN

## Insight into Interworkings

If using this program as blackbox, here are a few things to consider about the interworkings:

### swsc Window Types

swsc consides three window types starting at `v5.0.0`:

+ Candidate windows (size of `minWin`)
+ Extended candidate windows (size of `minWin*2`, extended `minWin/2` in both directions)
+ Window covering all candidates (size between `minWin*candidates` and UCE length)

### Change `minWin` and `candidates`

The default settings for these values are provided as a rough guide to realistic values, but are not meant to be the values used for all runs.

**For best results, `minWin*candidates` should be roughly `1/3` of the smallest UCE, indicating candidates can span the full length of the UCE.**

## Usage

### Installation

1. Install the Go language following instructions at: <https://golang.org/>
2. Run `go get -u bitbucket.org/rhagenson/swsc/...`
3. Run `swsc` by either:
    + Calling it directly if you added`$GOPATH/bin/` to your `$PATH`
    + Navigating to `$GOPATH/src/bitbucket.org/rhagenson/swsc/` and running `go build main.go && ./swsc`

### Running

Both `input`,`output`, and at least one metric (`--gc` or `--entropy`) must be set. See `swsc --help` for details.

### Reporting Errors

If you have found an error, or this tools does not work for you, please create an issue at: <https://bitbucket.org/RHagenson/swsc/issues> with details on when the error occurred, what the error states, and what was expected to occur, if known.

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

## Versions

A quick explanation of versions:

+ `v1.0.0`: Does a brute force search considering all possible windows `minWin` and up
+ `v2.0.0`: Uses candidate windows plus extension procedure (optimize large alignment performance)
+ `v3.0.0`: Candidate windows plus extension while using a single reference alignment (better performance)
+ `v4.0.0`: Candidate windows plus extension, single reference alignment, and remove redundant calculations
+ `v5.0.0`: Multiple candidate windows plus extension and single reference alignment
+ `v5.1.0`: Same as `v5.0.0`, but done in parallel for each UCE
+ `v6.1.0`: Update CLI to allow Nexus or FASTA+UCE csv input

Use `git checkout <version>` to move to a particular version (and `git checkout master` to move to the latest untagged development version). From there you can run either `go install` to install the particular version in `GOPATH` (overwritting any previous installed version) or `go build [-o <build name>]` to build the version in the current directory.

Versions can give different results so I would recommend using `v1.0.0` if you want the absolute best result (and have the time to wait for it to run a long, long time) or `v6.1.0` with a realistic `-minWin` and `-candidates` settings (rule of thumb: `minWin*candidates` should be roughly `1/3` of the smallest UCE, indicating candidates can span the full length of the smallest UCE).
