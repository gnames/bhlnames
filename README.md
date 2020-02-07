# bhlnames
bhlnames takes one or more scientific names and finds their usage in Biodiversity Heritage Library

## Introduction

[Biodiversity Heritage Library (BHL)][bhl] contains more than 200 000 volumes
(books, scientific journals, diaries of explorers, etc.) that contain some
important biodiversity information. Since the middle of 18th century, scientists
use Latinized scientific names as identifiers for known species. For biologists,
it is crucial to get information about species in publications. This program takes a
scientific name or a list of scientific names and returns references found
in BHL about the name. Quite often, species names change with time. In such
cases, we need to find information attached to older names as well as currently
accepted by community. The `bhlnames` uses [Catalogue Of Life (CoL)][col] synonymy
information to find publications not only about a given name but also about its
synonyms.

## Prerequisites

To be able to use this program you need

* a modern computer (laptop or desktop)
* one of the 3 operating systems (Linux, Mac OS, Windows)
* a functional Postgresql database
* 30+ GB of space on a hard drive
* 8GB or more of memory

## Installation

1. Download the [latest release] of `bhlnames`, untar or unzip the executable
`bhlnames` or `bhlnames.exe` and place it somewhere in your PATH.

2. Create a database (for example, `bhlnames`) on your Postgresql server. We
are not covering how to use Postgresql in this document. There are many
tutorials about it on the web. Make sure that the database is accessible from
the computer where you installed `bhlnames` executable.

3. Copy [`.bhlnames.yaml`][config] file to your HOME directory, modify database
parameters according to your settings. It is also important to setup an
`InputDir` directory for downloaded and temporary files, as well as for
key-value store databases.  You can leave other parameters as is for now.

The system should be ready for the initialization step now.

To make sure you have the right version of `bhlnames` run:

```bash
bhlnames -V
```

## Initialization

This step downloads all the needed BHL and names metadata on your computer.
Some of the data go to Postgresql database, others to a key-value store. You
do not need to worry about the creation of tables, or key-value databases; they
will be populated automatically. The program uses a file containing metadata
dump from BHL, as well as a remote `bhlindex` service. Dump provides
information about papers (parts in BHL terminology), volumes (issues), and
books (titles). The `bhlindex` provides fresh information about names and
their occurrences in BHL.

To start the initialization process type:

```bash
bhlnames init
```

The whole process will take about 3 hours, but it could take significantly
longer if your computer or internet connection is slow.

If for some reason you have to restart the program, you do not need to delete
working directories or the database. All of them will be updated automatically.
Some slow steps will not be repeated (such  as downloading full dump of BHL
database), unless you use `-r` option during initialization:

```bash
bhlnames init -r
```

In this case, everything will start from the beginning. In case if you
downloaded a newer version of `bhlnames` instead of older ones, you might need
to drop the database and create it again, and do init from scratch. Note that the BHL
dump is updated regularly, and it is good to rebuild your metadata set from
time to time from scratch.

## Usage

To find references for one name:

```bash
bhlnames refs "Pardosa moesta"
```

The result (in JSON format) will be sent to STDOUT and can be redirected to a
file

```bash
bhlnames refs "Pardosa moesta" > pm.json
```

By default, JSON is returned in its compact form. Optionally you can get data in
a more human-readable form with:

```bash
bhlnames refs "Pardosa moesta" -f pretty
```
You can also use [jq] or a similar tool

```bash
bhlnames refs "Pardosa moesta" | jq
```

In case if it is preferable to have the oldest publications last, you can reverse
sorting order with:

```bash
bhlnames refs "Pardosa moesta" -f pretty -d
```

To search for a large collection of names provide the name of a file instead (one
name per line):

```bash
bhlnames refs names.txt
```

For computers with modern multicore CPU, you can increase number of parallel
jobs. Usually, there is no much gain to go over 8 jobs.

```bash
bhlnames refs names.txt -j 8
```

To get a short version of data vithout details for references:

```bash
bhlnames refs names.txt -s
```

To get results without synonyms:

```bash
bhlnames refs names.txt -n
```

On a 12-core laptop, processing of 10000 names took about 40 seconds with 8
parallel jobs, and 2m 45sec with a single job. 10000 names generated 120MB of
results.

When you find the optimal number of jobs for your computer you can modify
`JobsNum` parameter in your version of the [`.bhlnames.yaml`][config] file
accordingly.

## Explanation of received data.

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

[bhl]: https://www.biodiversitylibrary.org/
[col]: https://www.catalogueoflife.org/col/
[latest release]: https://github.com/gnames/bhlnames/releases/latest
[config]: https://raw.githubusercontent.com/gnames/bhlnames/master/config_example/.bhlnames.yaml
[jq]: https://github.com/stedolan/jq
