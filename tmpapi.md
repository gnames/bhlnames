# API planning

## Data Sources

1. BHL metadata about titles, items, parts, pages
2. BHLindex found names data
3. BHLindex verification against CoL data
4. CoL name/nomeclatural reference data (~33% of CoL names have a nomenclatural reference)
5. Precalculated links betwen CoL nomenclatural references and BHL pages

## What possibilities we have

I. Get appearences of a name in BHL

* name (string)
* no-synonyms (bool) true

II. Get appearences of a taxon in BHL

* name (string)
* no-synonyms (bool) false

III. Find possible nomenclatural reference for a name (according to CoL)

* name (string)

IV. Find possible nomenclatural references of species for a higher taxon (CoL)

* taxon (string)

V. Find possible nomenclatural link for a name+reference

* name
* reference

(Using nomenclature-related scoring like sp.nov. etc.)

VI. Find possible BHL reference link for a name+reference

* name
* reference

(Ignoring nomenclature-related scoring like sp.nov. etc.)

<<<<<<< Updated upstream
VII. Find BHL links to items/parts most related to a higher taxon
=======
VII. Find BHL links to items/parts in the context of a taxon

Context here means a taxon of a particular rank with the highest
percentage of names for an item.
>>>>>>> Stashed changes

* taxon

Sort by "percentage,names-number" or "names-number,percentage"

VIII. Provide statistics about taxonomic context of BHL items/parts

* rank

The lower the rank, the more unrelated results we might get (should
correlate with percentage)

----------------------------------------------------------------------------

## Internal interfaces

I,II: Name(name Input, withSynonyms bool) []*namerefs.NameRefs

III: NameCoL(name Input, bestRes bool) []*namerefs.NameRefs

IV: TaxonCol(taxonInput, bestRes bool) []*namerefs.NameRefs

V,VI: NameRef(nameRef Input, isNomen bool, bestRes bool) []*namerefs.NameRefs

VII: TaxonItems(taxonInput, sortBy) []Items

VII: TaxonParts(taxonInput, sortBy) []Part

VIII: StatsItems(rankInput, sortBy) []Items

VIII: StatsParts(rankInput, sortBy) []Parts

----------------------------------------------------------------------------

## CLI API

----------------------------------------------------------------------------

## REST API

<!-- should I,II,V,VI be one resource? -->
I,II: /bhlnames.org/api/v0/bhl-links-by-name
V,VI: /bhlnames.org/api/v0/bhl-links-by-name-ref

<!-- should III,IV be one resource? -->
III:  /bhlnames.org/api/v0/bhl-nomen-links-by-name
IV:   /bhlnames.org/api/v0/bhl-nomen-links-by-taxon

<!-- these take higher taxon name, return items or pages -->
VII:  /bhlnames.org/api/v0/items
VIII: /bhlnames.org/api/v0/parts

VII:  /bhlnames.org/api/v0/items-stats
VIII: /bhlnames.org/api/v0/parts-stats
