# Changelog

# v0.0.7
* Add many, inneJoin, nonResult flags to meta info.
* Add checking that subquery on select clause is not allowed with aggregation query.
* Edit README.

# v0.0.6
* Add used functions, parameters, datetimeLiterals to meta info.
* Set max query depth to meta info.
* Add elapsed time to meta info.
* Add query graph to meta info.
* Add query id to ViewGraph.
* [FIX] Fix query depth.

# v0.0.5
* Add Depth, QueryDepth, Query to QueryGraph.
* Add view graph to query meta info.
* Add object max depth to query meta info.
* Set parent view id to object,
* Set query id to query.

# v0.0.4
* Add meta information to query object.
* [FIX] Fix fieldset parameter resolution.

# v0.0.3
* Register unregistered fields in the correlated subquery.
* Set ViewId to objects and fields.
* Add colIndex to SoqlFieldInfo.
* [FIX] Fix co-related subquery checking.
* [FIX] Subquery fields should be normalized after building per-object infos.
* [FIX] Fix field checking: item must be included in a Group By clause.
* Add and edit TODO comments.
* Edit wasm demo.

# v0.0.2
* Edit README.

# v0.0.1
* First release.
