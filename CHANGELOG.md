# Changelog

## Unreleased

## [v0.0.6]

- Add [#20] run bhlnames against all CoL files.
- Add [#16] functional REST API.
- Add [#14] refactor initialization code to use interfaces.
- Add [#13] refactor reference code to use interfaces.
- Add [#12] link nomenclatural events to BHL.
- Add [#11] use nomenclatural annotations in output and collection of reference
            data.
- Fix [#19] memory leak in postgres database.

## [v0.0.5]

- Add list of detected synonyms for the summary.
- Add images url from Google Images

## [v0.0.4]

- Add [#10] currently accepted canonical in output. Add option for
            for a short output.
- Add [#8] option to return result without synonyms.
- Add [#7] add documentation.
- Fix [#9] close DB connections in the end of Refs methods.

## [v0.0.3]

- Add [#5]: return titles, items, parts for a name or a list of names
            in chronological order.
            The application is usable at this poinit. So we make
            the first binary release.
- Add [#6]: get part (paper) data out of `page_id`.

## [v0.0.2]

- Add [#4]: upload data for name-strings and their occurrences to db.

## [v0.0.1]

- Add [#3]: import BHL metadata to database
- Add [#2]: set migrations for database.
- Add [#1]: return version number.

## Footnotes

This document follows [changelog guidelines]

[v0.0.6]: https://github.com/gnames/bhlnames/compare/v0.0.5...v0.0.6
[v0.0.5]: https://github.com/gnames/bhlnames/compare/v0.0.4...v0.0.5
[v0.0.4]: https://github.com/gnames/bhlnames/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/gnames/bhlnames/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/gnames/bhlnames/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/gnames/bhlnames/compare/v0.0.0...v0.0.1

[#20]: https://github.com/gnames/bhlnames/issues/20
[#19]: https://github.com/gnames/bhlnames/issues/19
[#18]: https://github.com/gnames/bhlnames/issues/18
[#17]: https://github.com/gnames/bhlnames/issues/17
[#16]: https://github.com/gnames/bhlnames/issues/16
[#15]: https://github.com/gnames/bhlnames/issues/15
[#14]: https://github.com/gnames/bhlnames/issues/14
[#13]: https://github.com/gnames/bhlnames/issues/13
[#12]: https://github.com/gnames/bhlnames/issues/12
[#11]: https://github.com/gnames/bhlnames/issues/11
[#10]: https://github.com/gnames/bhlnames/issues/10
[#9]: https://github.com/gnames/bhlnames/issues/9
[#8]: https://github.com/gnames/bhlnames/issues/8
[#7]: https://github.com/gnames/bhlnames/issues/7
[#6]: https://github.com/gnames/bhlnames/issues/6
[#5]: https://github.com/gnames/bhlnames/issues/5
[#4]: https://github.com/gnames/bhlnames/issues/4
[#3]: https://github.com/gnames/bhlnames/issues/3
[#2]: https://github.com/gnames/bhlnames/issues/2
[#1]: https://github.com/gnames/bhlnames/issues/1

[changelog guidelines]: https://github.com/olivierlacan/keep-a-changelog
