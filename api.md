/names
  input:
    name (mandatory)
    reference (optional)
    withSynonyms (true of false as default?)
    data-source (CoL is default)
    is-nomenlatural-event (false as default)
    offset (default 0)
    limit (from 1 to maxAllowed)

  output:
    input
    matched name
    synonyms
    BHLmetadata:
      number of results
      number of occurrences
      minYear
      maxYear
      minScore
      maxScore

/pages
  input:
    the same
  output:
    same as name
    +
    array of
      metadata about page
      score

/items
  input:
    higher taxon
  output:
    items where the taxon is the main.

/parts
  input:
    higher taxon
  output:
    items where the taxon is the main.

/stats
  input:
    item-id optionsl
    part-id optional
    offset 0 default
    limit from 1 to max (max is default)


Tables names

cache_page_results
cache_page_all_results
