# Concept: DonutForest

A *DonutForest* is so named under two metaphors. It is both a "Donut" and a "Forest".

### Donut
As a donut, it is the bottom item of a cake-tier-donut metaphor. A donut is one of the concentric rings that comprises
a tier of a cake. When an upper tier fills, the entire tier is re-baked into a new outer donut of the tier below.

### Forest
As a forest, it is the top item of a forest-tree-leaf metaphor. A forest is a collection of trees. The forest
contains 256^n trees, where n is the tier index. `TierBelow[0]` has forests containing just one tree each.
`TierBelow[1]` has forests containing 256 trees each, and so on. Trees in a forest are found by an index in a
"jump table" of 256^n entries. The index is enumerated by the first n bytes of a hash being considered.

## Hash lookup
When looking up a hash, once the relevant tree is found based on the first n bytes of the hash, forks
in the tree ("nodes") are traversed to find the leaf containing the hash. Each node specifies a (non-sequential)
byte of the hash to examine, and 256 arcs to follow in the case of each possible value of that byte.
The arcs ("slots") each lead to another node, which may be a leaf.

If a slot does not specify a node, the hash is not found in the forest.

If a slot specifies a leaf node, enough bytes of the hash have been examined to identify an entry
within the forest that matches the hash uniquely within the forest. Further (possibly all) bytes of
the hash may be examined to confirm the entry matches the entire hash.

A given hash may exist in any one of the DonutForest's in any of the tiers of the cake. In a sense
all DonutForest's should be examined in parallel to find the hash, though on finding an exact match
the searches in other DonutForest's may be abandoned.

## Storage by tree level
A DonutForest is represented on disk in a distributed way in three places.
(Ultimately, all the nodes of all the trees of all the DonutForest's of a tier are stored in
multiple files `Tier<n>/Level<LL>Nodes.bin`, where <LL> is the level of the tree, but we'll get back to that...)

* Firstly (non-distributed), a DonutForest has an entry in `Tier<n>/DonutForestsInfo.bin`.
    * For each level LL, this entry points to contiguous "indexBytes" and "nodeBytes" within `Level<LL>Nodes.bin`.

* Secondly (non-distributed), a DonutForest has an entry in `Tier<n>/DonutForestsJumpTables.bin`.
    * This entry is a list of 256^n (possibly just one!) nodeId's of the roots of the 256^n trees in the forest.

* Thirdly (distributed), a DonutForest has a pair of entries in each `Tier<n>/Level<LL>Nodes.bin`.
    * The first entry "indexBytes" describes the "designed" set of formats used to describe the nodes in the second entry.
    * The second entry "nodeBytes" contains the actual nodes of the forest, in the formats described by the first entry.

## Design commentary
At first sight this would appear to be overcomplicated.
There are two reasons for this design.
1) By "designing" a set of formats for a DonutForest within a Level, we skip the need for an extensive
   nodeId to byte number lookup table
2) By distributing nodes by level, frequently accessed levels are kept "close together" on disk;
   this allows us to mmap the files and let the OS cache the files more effectively.
