# package weddingcakefront
WeddingCake is an LSM (Log-Structured Merge Tree) for indexing a chronologically presented sequence of cryptographic
hashes. After injecting some hashes, you can (trivially, forward lookup) request the hash that was presented
at a given chronological index. You can also (reverse lookup) request the chronological index that a given hash
was presented at.
* *weddingcakefront* is the high level package that a programmer would use
* *weddingcakeback* is the low level package that WeddingCake uses internally
## Tiered cake metaphor
TierTop is the top tier of the wedding cake, representing the most recent and possibly most frequently accessed data.
As time progresses, older data is moved to lower tiers (0+), allowing for efficient retrieval and storage.
We build the tiered cake from the top down.
### TierTop is different
TierTop is different, being primarily implemented as an in-memory map. Hashes are only ever presented to TierTop.
### Tiers 0+ comprised of up to 255 DonutForest's
Under TierTop is Tier 0.
Each tier 0+ is comprised of zero or more concentric DonutForest's (they are Donuts from the high level
cake-tier-donut metaphor, but they are also Forests from a lower level forest-tree-leaf perspective.)
### A whole tier is "baked" into a single DonutForest
When a tier reaches a certain size, it is "baked" into a single DonutForest in the tier below.
* TierTop is "full" when it has been presented with 65535 hashes (including those that were duplicates of prior hashes). 
    * All hashes are baked into a single DonutForest in tier 0.
    * TierTop then becomes empty, ready for more hash injections
* Tier 0 is "full" when it contains 255 DonutForests.
    * Hashes from ALL 255 of tier 0's DonutForests are baked into a single DonutForest in tier 1.
    * Consequently, DonutForest's are much bigger in tier 1 compared to tier 0.
    * Tier 0 then becomes empty.
* Similarly, tiers 1+ are "full" when they contain 255 DonutForests.
## Limited support for duplicate hashes
* Duplicate hashes are "tolerated"
* Forward lookup (index to hash) correctly returns the hash, even if the same hash was presented multiple times.
* Reverse lookup (hash to index) correctly returns the FIRST index at which the hash was presented.
* Reverse lookup (hash to index) gives NO INFORMATION about whether the hash was presented multiple times.
