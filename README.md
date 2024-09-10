# bhlnames

bhlnames takes one or more scientific names and finds their occurrences in
Biodiversity Heritage Library. The app also creates an apparent link between a
name-string/publication input and the corresponding BHL reference.

## Introduction

[Biodiversity Heritage Library (BHL)][bhl] contains more than 250 000 volumes
(books, scientific journals, diaries of explorers, etc.). BHL provides an
important biodiversity information. Since the middle of 18th century,
scientists use Latinized scientific names as identifiers for known species. For
biologists, it is crucial to get information about species in publications,
it is especially valuable to find original description of species or creation
of a new binomial genus/species combination.

This program tries to answer the following questions:

- Where a particular name-string appears in BHL?

- Given a names string, where species it assigned to appears in BHL? In this
  case we find all synonyms of the name-string and the currently accepted
  name of the species

- Given a name-string and a reference to its original publication, does this
  publication exists in BHL, and what is the link to it?

- Given a name-string and a reference to a publication about the name-string,
  find a link to this publication in BHL.

The `bhlnames` app uses [Catalogue Of Life (CoL)][col] synonymy information to
find publications not only about a given name but also about its synonyms. In
the future it will support other resources with synonymy information.

## Provided functionalities

1. Find references in BHL where a scientific name-string appears.

   Searching for a name-string without synonyms.

2. Find references in BHL where a taxon, represented by a name-string appears.

   Determining the corresponding taxon an entered name-string. This taxon
   has a currently accepted name and a variety of synonym name-strings.
   Entered name-string might match either a synonym, or the currently accepted
   name. Searching for references where any of the found name-strings
   appear.

3. Find a reference to the original description of a name or a new combination.

   In this case the input is a nomenclatural event represented by a name and
   its official nomenclatural publication. We use provided information trying
   to find a BHL reference that corresponds to that publication.

4. Find a link to a publication, using a publication reference and
   a name used in the publication.

5. Provide REST API and command line tool to access aforementioned
   functionalities.

## Prerequisites

To be able to use this program you need

- a modern computer (laptop or desktop)
- one of the 3 operating systems (Linux, Mac OS, Windows)
- a PostgreSQL database
- 30+ GB of space on a hard drive
- 32GB or more of memory

## Installation

1. Download the [latest release] of `bhlnames`, untar or unzip the executable
   `bhlnames` or `bhlnames.exe` and place it somewhere in your PATH.

2. Create a database (default name is `bhlnames`) on your Postgresql server. We
   are not covering how to use Postgresql in this document. There are many
   tutorials about it on the web. Make sure that the database is accessible
   from the computer where you installed `bhlnames` executable.

3. Run `bhlnames -V`. It will show the version number of the program and will
   create a configuration file. Terminal output will provide information where
   the file is located. For example on Linux machines it will be placed in
   `$HOME/.config/bhlnames.yaml`.

4. Go to the configuration file and modify database parameters in the config.
   You can also change setup for `InputDir` directory for downloaded and
   temporary files, as well as for key-value store databases. You can leave
   other parameters as is for now.

The system should be ready for the initialization step.

## Initialization

This step downloads all the needed BHL and names metadata on your computer.
Some of the data go to the Postgresql database, others to a key-value store.
You do not need to worry about the creation of tables, or key-value databases;
they will be populated automatically. The program uses a file containing
metadata dump from BHL, as well as names data created by `bhlindex`. Dump
provides information about papers (parts in BHL terminology), volumes (issues),
and books/journal volume sets (titles). The `bhlindex` provides fresh
information about names and their occurrences in BHL.

To start the initialization process type:

```bash
bhlnames init
```

The whole process will take about 1-2 hours, but it could take significantly
longer if your computer or internet connection are slow.

If for some reason you have to restart the program, you do not need to delete
working directories or the database. All of them will be updated automatically.
Some slow steps will not be repeated (such as downloading full dump of BHL
database), unless you use `-r` option during initialization:

```bash
bhlnames init -r
```

In this case, everything will start from the beginning. In case if you
downloaded a newer version of `bhlnames` instead of older ones, you might need
to drop the database and create it again, and do init from scratch. Note that
the BHL dump is updated regularly, and it is good to rebuild your metadata set
from time to time from scratch.

## Usage

To find references to a whole taxon (synonyms and currently accepted name)
from a name-string:

```bash
bhlnames name "Pardosa moesta"
```

The result (in JSON format) will be sent to `STDOUT` and can be redirected to a
file

```bash
bhlnames name "Pardosa moesta" > pm.json
```

By default, JSON is returned in its compact form. Optionally you can get data in
a more human-readable form with:

```bash
bhlnames name "Pardosa moesta" -f pretty
```

You can also use [jq] or a similar tool

```bash
bhlnames name "Pardosa moesta" 2>/dev/null | jq
```

In case if it is preferable to have the oldest publications last,
you can reverse sorting order with:

```bash
bhlnames name "Pardosa moesta" -f pretty -d
```

To search for a large collection of names provide the name of a file instead
(one name per line):

```bash
bhlnames name names.txt
```

For computers with modern multi-core CPU, you can increase number of parallel
jobs. Usually, there is no much gain from setting more than 8 jobs.

```bash
bhlnames name names.txt -j 8
```

To get a short version of data without details for references:

```bash
bhlnames name names.txt -s
```

To get results without synonyms:

```bash
bhlnames name names.txt --no_synonyms
```

To find a link to a name-string with its original reference you can use a
CSV.

Mimumal fields are:

```csv
Id,NameString,RefString
```

Where `Id` is an internal identifier, `NameString` is a full scientific name
with authorship and year (if they are given), `RefString` is an unparsed
reference string.

If you have a more detailed information the following fields can also be used:

`NameCanonical` canonical form of a name, `NameYear` the year of the name
publication, `RefYear` the year of the reference publication.

Most

You can use the following command:

```bash
bhlnames nomen name-refs.csv
```

The result will be send to `STDOUT` in a compact JSON format, one datum per
line. Use [jq] or similar program to render 'pretty' version of JSON.

On a 12-core laptop, processing of 10000 names took about 40 seconds with 8
parallel jobs, and 2m 45sec with a single job. 10000 names generated 120MB of
results.

When you find the optimal number of jobs for your computer you can modify
`JobsNum` parameter in your version of the [`bhlnames.yaml`][config] file
accordingly.

## REST API

To start `bhlnames` as a server on a port 1234:

```bash
bhlnames rest -p 1234
```

### REST end-points

- `/name_refs` (POST) to find occurrences of a names-string. Takes JSON encoded
  list of name-strings as an argument.

- `/taxon_refs` (POST) to find occurrences of a taxon. Takes JSON encoded list
  of name-strings as an argument.

- `/nomen_refs` (POST) to find a link to the provided reference.
  Takes a JSON-encoded structure.

For more details how to use API you can refer to the [REST test file].

## Explanation of received data

### Taxon and name-string output

This output is created by `bhlnames name` command.

Returning information can be quite large. You will get data in chronological
order. If there is data about a specific paper, we return information about
the paper; if we only have information about the item, or title, we return that
information.

For every item, we return the most populated Linnean kingdom for the
unique names found in the item. We also provide the percentage of names that
got resolved to that kingdom. We use the managerial CoL classification for this
purpose.

In addition, we return a "context" of the item. The context is the lowest taxon
that still contains at least 50% of all names found in the item. This gives a
better idea about the item's biological content. For example, if the context is
"Araneae" the item is mostly about spiders.

You can use provided `page_id` and `item_id` to find information on BHL
website. For example to find `page_id` 26895127 use:

`https://www.biodiversitylibrary.org/page/26895127`

### Original reference output

This output is created by `bhlnames nomen` command.

You receive input information (id, name, reference) as well as 0 or 1 reference
from BHL that was the picked as the best candidate to the link for the
input reference. If we did not get any feasible candidates, no BHL reference
is provided.

TODO: Explain score and score thresholds.

## Development

### Running tests

1. Install `direnv`, copy `envrc.example` to `.envrc`.
2. Modify `.envrc` according to your needs, enable `direnv`.
3. Install bhlnames and its data as described above.
4. in `./internal/io/restio/restio.go` go to TODO and swap development
   and production servers. (Dont forget to swap them back for production)
5. Start RESTful service in another terminal window:

   ```bash
   bhlnames rest
   ```

6. Run `make tools`
7. Run `make test`
8. Run `make testrest`

## Authors

- [Dmitry Mozzherin], [Geoff Ower]

## License

Released under [MIT license]

[bhl]: https://www.biodiversitylibrary.org/
[col]: https://www.catalogueoflife.org/col/
[latest release]: https://github.com/gnames/bhlnames/releases/latest
[config]: https://raw.githubusercontent.com/gnames/bhlnames/master/config_example/.bhlnames.yaml
[jq]: https://github.com/stedolan/jq
[dmitry mozzherin]: https://github.com/dimus
[geoff ower]: https://github.com/gdower
[mit license]: https://github.com/gnames/bhlnames/blob/master/LICENSE
[rest test file]: https://github.com/gnames/bhlnames/blob/master/rest/rest_test.go
