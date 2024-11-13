### Release Update (2024.11.11)

#### NNV-Edge

[Please check the detailed results and changes.](./examples/release/2024_11_11_release.md)

- The Edge version has been released first (currently still in the RC stage).
- For detailed information on the Edge version, please refer to the **[Edge section]**.
- The Edge version is written entirely in pure Go.
- It includes F16 quantization.
- Due to the nature of the Edge version, several conveniences have been removed, requiring more adjustments from the user.

#### NNV

- With the removal of Usearch, the CGO dependency is also eliminated.
- We are reviewing speed and accuracy while revising the existing HNSW.
- Development may slow significantly until the project is transitioned to using Edge.

### Release Update (2024.11.10)

- We intended to incorporate usearch for its enhanced HNSW capabilities. However, the documentation is still immature, there are critical aspects that prevent direct error handling, and there is a lack of C++ domain expertise to enable the developer to modify and update it directly. Therefore, we need to develop a more mature HNSW than the existing implementation.

### Release Update (2024.11.02)

[Please also review the test results.](./examples/release/2024_11_02_release.md)

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
