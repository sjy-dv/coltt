## 🎉 Release Update - 2024.12.09

### 🔹 NNV-Edge

- **Planned Work for Enhancing Edge Performance**: During the current core development, we have achieved very fast write and read operations through sharding methods. We plan to add this sharding logic to the edge to expect speed improvements on the edge and to address existing performance enhancements.

---

### 🔹 NNV

- **HNSW Test Completed**: Achieved 0.87 milliseconds in searching 1 million vectors. It is 0.87 milliseconds, not seconds (second is 0.00087 seconds). This is a very gratifying achievement.
- **Progress on PQ and BQ**: Continuous review of PQ and BQ is underway.
- **Integration of Existing Quantization**: Planning to proceed with quantization integration (Report work is delayed due to a heavy workload. 😢)

---

## 🎉 Release Update - 2024.11.20

### 🔹 NNV-Edge

#### Enhancements

- **Edge Performance Improvement**: Enhanced the performance of Edge.
- **Decreased Storage Requirement**: Reduced the storage needed for 1,000,000 128-dimensional vectors from **2.5GB** to **1.35GB** (Milvus: **1.46GB**).
- **Average Search Time** for 1,000,000 128-dimensional vectors: **0.22 sec** (Milvus: **0.04 sec**).
- **Continuous Performance Enhancement**:
  - Clearly identified points for ongoing performance improvements.
  - Discovered that efficient parallel search using a worker pool and iterating over the vector space as an array significantly boosts performance.
  - These enhancements require rewriting specific code sections and are planned for long-term development.
  - Retrieving all data from disk introduces overhead; therefore, combining disk access with memory usage is necessary.
- **Data Storage**: Data is stored on disk with only the flush operation for indexing remaining.
- **Safe Recovery**: Even if the server crashes, indexes and all data can be safely recovered.

---

### 🔹 NNV

- **README Revamp**: The README will be updated soon.
- **Improved HNSW Development**: A more advanced HNSW algorithm is planned for development.
- **Continued Testing of PQ**: PQ (Product Quantization) is continuously being tested. According to Weaviate's documentation, PQ performs very poorly on datasets with tens of thousands of data points. I am experiencing this myself. (https://weaviate.io/blog/pq-rescoring)
- **Development of BQ (Binary Quantizer)**: Development of BQ is scheduled.
- **Quantization of HNSW with F8, F16, BF16**: In this development phase, we are also considering quantizing HNSW with F8, F16, and BF16. (Applying this is not difficult; the challenge is figuring out how BQ and PQ will compete.)

## 🎉 Release Update - 2024.11.14

### 🔹 NNV-Edge

- No updates in this release.

---

### 🔹 NNV

See the [detailed comparison with ChromaDB for search results](./examples/release/2024_11_14_release.md).

#### Enhancements

- **Restored HNSW**: The previously used HNSW algorithm has been reintroduced.
- **New Product Quantization (PQ)**: HNSW Product Quantization has been added for improved efficiency.
- **Pure Go Implementation**: All CGO dependencies have been removed, making the implementation entirely in Go.
- **Optimized Search Speed**:
  - 50,000-item dataset: **< 14ms**
  - 10,000-item dataset or fewer: **< 3ms**

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
