### Release Update (2024.11.10)

- We intended to incorporate usearch for its enhanced HNSW capabilities. However, the documentation is still immature, there are critical aspects that prevent direct error handling, and there is a lack of C++ domain expertise to enable the developer to modify and update it directly. Therefore, we need to develop a more mature HNSW than the existing implementation.

### Release Update (2024.11.02)

[Please also review the test results.](./examples/2024_11_02_release.md)

- Automatic indexing is supported.
  - Limitations
    - All metadata is indexed without user configuration.
    - RAM is required to maintain the index.
    - Support for various operators is limited; currently, only matching is possible.
- Searches can only be performed using filters.
- Hybrid search is supported.
  - Limitations
    - All filters must be processed as strings. For example, if age is stored as an integer (20), it must be searched using "age": "20".
    - However, this is only a restriction during searches; the actual data retains its original type.

### Release Update (2024.10.31)

- Fast cache data storage and loading is supported through INFLATE/DEFLATE.
- nnlogdb has stabilized, and log storage and bucket-based partitioning have been completed.
