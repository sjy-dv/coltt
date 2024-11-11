# NNV (No-Named.V)

# ⚠️Currently, the project has become more unstable due to migrating to FastHnsw.

![logo](./examples/assets/logo.png)

NNV (No-Named.V) is a database designed to be implemented from scratch to production. NNV can be deployed in edge environments and used in small-scale production settings. Through the innovative architectural approach described below, it is envisioned and developed to be used reliably in large-scale production environments as well.

### Release Update (2024.11.11) [UPDATE HISTORY](./UPDATE-LOG.md)

#### NNV-Edge

[Please check the detailed results and changes.](./examples/2024_11_11_release.md)

- The Edge version has been released first (currently still in the RC stage).
- For detailed information on the Edge version, please refer to the **[Edge section]**.
- The Edge version is written entirely in pure Go.
- It includes F16 quantization.
- Due to the nature of the Edge version, several conveniences have been removed, requiring more adjustments from the user.

#### NNV

- With the removal of Usearch, the CGO dependency is also eliminated.
- We are reviewing speed and accuracy while revising the existing HNSW.
- Development may slow significantly until the project is transitioned to using Edge.

### Update Preview

**⚠️I don't know when it will be updated. (I just want to develop 😭)**

#### NNV-Edge

- I need to add detailed logging.
- I’m planning to work on a project using Edge and will improve any areas that need adjustments based on progress and feedback.

#### NNV

- Quantization functionality will be added.
- HNSW will be faster.
- More vector distance algorithms will be added.
- CGO dependency will be introduced. => The introduction of Usearch has failed, and as a result, the CGO dependency is planned to be removed.
- Fast in-memory storage and reliable disk storage will be added.
- A process to periodically save data must be added after checking the system's idle state.
- An automatic recovery feature must be added.
- Expressions, such as various range searches, must be supported in filters.
- Benchmarking must be conducted after the system stabilizes.
- A load balancer must be developed after the system stabilizes.

### ⚠️ Warning

- It may be slow because you are not currently focused on this task.

### Run from the source code.

```sh
Windows & Linux
git clone https://github.com/sjy-dv/nnv
cd nnv
go run cmd/root/main.go

MacOS
**The CPU acceleration (SSE, AVX2, AVX-512) code has caused an error where it does not function on Mac, and it is not a priority to address at this time.**

git clone https://github.com/sjy-dv/nnv
cd nnv
source .env
deploy
make simple-docker
```

# Index

- [Features](#features)
- [ARCHITECTURE](#architecture)

  - [LoadBalancer&DatabaseIntegration](#loadbalancer--database-integration)
  - [JetStream(Nats)Multi-Leader](#jetstreamnats-multi-leader)
  - [InternalDataFlow](#i-will-explain-the-internal-data-storage-flow)
  - [cache-data-is-safe?](#disk-files-can-sometimes-become-corrupted-and-fail-to-open-leading-to-significant-issues-is-cached-data-safe)
  - [Edge](#what-is-nnv-edge)

- [BugFix](#-bugfix)

## Features

When planning this project, I gave it a lot of thought.

When setting up the cluster environment, it's natural for most developers to choose the RAFT algorithm, as I had always done before. The reason being that it's a proven approach used by successful projects.

However, I began to wonder: isn't it a bit complex? RAFT increases read availability but decreases write availability. So, how would I solve this if multi-write becomes necessary in the long run?

Given the nature of vector databases, I assumed that most services would be structured around batch jobs rather than real-time writing. But does that mean I can just skip addressing the issue? I didn't think so. However, building a multi-leader setup on top of RAFT using something like gossip felt extremely complex and difficult.
![img1](./examples/assets/raft_problem.png)

Therefore, as of today (2024-10-20), I am considering two architectural approaches.

## ARCHITECTURE

The architecture is divided into two approaches.

### LoadBalancer & Database Integration

First, a load balancer is placed at the front, supporting both sharding and integration of the data. The internal database exists in a pure state.

| ![architecture1](./examples/assets/arch1.png) | ![architecture2](./examples/assets/arch2.png) |
| :-------------------------------------------: | :-------------------------------------------: |
|                  Replica LB                   |                   Shard LB                    |

The replication load balancer waits for all databases to successfully complete writes before committing or rolling back, while the shard load balancer distributes the load evenly across the shard databases to ensure similar storage capacities.

The key difference is that replication can slow down write operations but provides faster read performance in the medium to long term compared to the shard load balancer. On the other hand, the shard approach offers faster write speeds because it only commits to a specific shard, but reading requires gathering data from all shards, which is slower initially but could become faster than replication as the dataset grows.

Therefore, for managing large volumes of data, the shard balancer is slightly more recommended. However, the main point of both architectures is their simplicity in setup and management, making them as easy to handle as a typical backend server.
![arch1_structure](./examples/assets/arch1_struct.png)

### JetStream(Nats) Multi-Leader

![arch4](./examples/assets/arch4.png)

The second approach utilizes JetStream for the configuration.

While this is architecturally simpler than the previous approach, from the user's perspective, the setup is not significantly different from RAFT.

However, the key difference is that, unlike RAFT, it supports multi-write and multi-read configurations, rather than single-write and multi-read.

In this approach, the database is configured in a replication format, and JetStream is used to enable multi-leader configurations.

![arch5](./examples/assets/arch5.png)
Each database contains its own JetStream, and these JetStreams join the same group of topics and clusters. In this case, whenever all nodes attempt to publish changes to a row, they pass through the same JetStream. If two nodes attempt to modify the same data in parallel, they will compete to publish their changes. While it's possible to prevent changes from being propagated, this could lead to data loss. According to the RAFT quorum constraint in JetStream, only one writer can publish the change. Therefore, we designed the system to allow the last writer to win. This is not an issue for vector databases because, compared to traditional databases, the data structure is simpler (this doesn't imply that the system itself is simple, but rather that there are fewer complex transactions and procedures, such as transaction serialization). This also avoids global locks and performance bottlenecks.

![summary](./examples/assets/summary.png)

### Summary:

1. **RAFT and Quorum Constraints**  
   RAFT is an algorithm that dictates which server writes data first. In RAFT, the concept of a **quorum** refers to the minimum number of servers required to confirm data before it's written. This ensures that even if two servers try to write data simultaneously, RAFT allows only one server to write first.
2. **Last Writer Wins**  
   Even if one server writes data first, the server that writes last ultimately "wins." This means that the data from the last server to write will overwrite the previous server’s data.
3. **Transaction Serialization Concerns**  
   Transaction serialization refers to ensuring that consistent actions occur across multiple tables. In NNV, to improve performance, global locking (locking all servers before writing data) is avoided. Instead, when multiple servers modify data simultaneously, the last one to modify it will win. This approach is feasible because vector databases are simpler than traditional databases—they don’t require complex transaction serialization across multiple tables or collections.
4. **Why This Design?**  
   The primary reason is performance. Locking all servers before processing data is safe but slow. Instead, allowing each server to freely modify data and accepting the last modification as the final result is faster and more efficient.

### I will explain the internal data storage flow.

![arch6](./examples/assets/arch6.png)

First, HNSW operates in memory internally, and its data is stored as cached files. However, this poses a risk of data corruption in the event of an abnormal shutdown.

To address this, nnlogdb (no-named-tsdb) is internally deployed to track insert, update, and delete events. Since only metadata and vectors are needed without node links, this is not a significant issue.

The observer continuously compares the tracked log values with the latest node, and if a problem arises, HNSW recovery is initiated.

### Disk files can sometimes become corrupted and fail to open, leading to significant issues. Is cached data safe?

![arch7](./examples/assets/arch7.png)
Cache data files support fast loading and saving through the INFLATE/DEFLATE compression algorithm. However, cache files are inherently much less stable than disk files.

To address this, we deploy "old" versions. These versions are not user-specified; instead, they are managed internally. During idle periods, data changes are saved as new cache data, while the previous stable open file version is stored as the "old" version. When this happens, the last update time of the "old" version aligns with the sync time in nnlogdb.

To manage disk usage efficiently, all previous partitions up to the reliably synced period are deleted.

This approach ensures stable data management.

### Does disk storage also have structural flexibility?

Disk storage will not initially have structural flexibility. However, in the long term, we aim to either introduce flexibility for the disk structure or, unfortunately, impose some restrictions on the memory side. While no final decision has been made, we believe memory storage should maintain flexibility, so we’re likely to design disk storage with some degree of structural flexibility in the future. SQLite will be supported as the disk storage option.

#### What is NNV-Edge?

Edge refers to the ability to transmit and receive data on nearby devices without communication with a central server. However, in practice, "Edge" in software may sometimes differ from this concept, as it is often deployed in lighter, resource-constrained environments compared to a central server.

NNV-Edge is designed to operate quickly on smaller-scale vector datasets (up to 1 million vectors) in a lightweight manner, transferring automated tasks from the original NNV back to the user for greater control.

Advanced algorithms like HNSW, Faiss, and Annoy are excellent, but don’t you think they may be a bit heavy for smaller-scale specs? And setting aside algorithms, while projects like Milvus, Weaviate, and Qdrant are built by brilliant minds, aren’t they somewhat too resource-intensive to run alongside other software on small, portable devices?
![arch9](./examples/assets/arch9.png)
That’s where NNV-Edge comes in.

What if you distribute multiple edges? By using NNV-Edge with the previously mentioned load balancer, you can create an advanced setup that shards data across multiple edges and aggregates it seamlessly!
