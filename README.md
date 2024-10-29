# NNV (No-Named.V)

![logo](./examples/assets/logo.png)

NNV (No-Named.V) is a production database project by a developer aspiring to gain recognition. The project is designed as a KV database, aiming to support FLAT (already supported => cancel) and HNSW indexing in the long term. Bitmap-based indexing and quantization for vector indexes are supported (quantization already supported => cancel). Additionally, it aims to incorporate real-time streaming functionality to enable versatile use cases.

Additionally, its flexible and innovative cluster architecture presents a new vision.

### ⚠️ Warning

- ~~HNSW accuracy is lower than expected. Currently being edited.~~
- It may be slow because you are not currently focused on this task.
- The hybrid search method using bitmap indexing within metadata is scheduled to be added after the initial release.

# Index

- [Features](#features)
- [ARCHITECTURE](#architecture)

  - [LoadBalancer&DatabaseIntegration](#loadbalancer--database-integration)
  - [JetStream(Nats)Multi-Leader](#jetstreamnats-multi-leader)
  - [InternalDataFlow](#i-will-explain-the-internal-data-storage-flow)
  - [cache-data-is-safe?](#disk-files-can-sometimes-become-corrupted-and-fail-to-open-leading-to-significant-issues-is-cached-data-safe)

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

### 🐛 BugFix

```go
dataString := []string{
		"I usually eat a sandwich",
		"When I'm hungry, I tend to eat porridge",
		"I like fixing cars",
	}

	query := "I'm hungry, what should I eat?"

	addString := []string{
		"Someday, I'll become a great person",
		"I've thought about what to eat when I'm hungry, and cookies are definitely the best",
	}

	// create dataset
	originDataset := make([]*dataCoordinatorV1.ModifyDataset, 0, 3)
	for _, data := range dataString {
		vec, err := embeddings.TextEmbedding(data)
		if err != nil {
			log.Fatal(err)
		}
		Uid := uuid.New().String()
		metas, _ := msgpack.Marshal(map[string]interface{}{
			"description": data,
			"_id":         Uid,
		})
		originDataset = append(originDataset, &dataCoordinatorV1.ModifyDataset{
			BucketName: "tbucket",
			Id:         Uid,
			Vector:     vec,
			Metadata:   metas,
		})
	}
	afterDataset := make([]*dataCoordinatorV1.ModifyDataset, 0, 3)
	for _, data := range addString {
		vec, err := embeddings.TextEmbedding(data)
		if err != nil {
			log.Fatal(err)
		}
		Uid := uuid.New().String()
		metas, _ := msgpack.Marshal(map[string]interface{}{
			"description": data,
			"_id":         Uid,
		})
		afterDataset = append(afterDataset, &dataCoordinatorV1.ModifyDataset{
			BucketName: "tbucket",
			Id:         Uid,
			Vector:     vec,
			Metadata:   metas,
		})
	}
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Println(err)
	rclient := resourceCoordinatorV1.NewResourceCoordinatorClient(conn)

	res, err := rclient.CreateBucket(context.Background(), &resourceCoordinatorV1.Bucket{
		BucketName: "tbucket",
		Dim:        384,
		Space:      resourceCoordinatorV1.Space_Cosine,
	})
	fmt.Println(err, res.Status)

	dclient := dataCoordinatorV1.NewDatasetCoordinatorClient(conn)
	for _, data := range originDataset {
		res, err := dclient.Insert(context.Background(), data)
		fmt.Println(res.Status, res.Error, err)
	}
	// first search
	searchVec, _ := embeddings.TextEmbedding(query)
	fmt.Println("old data search")
	resp, err := dclient.Search(context.Background(), &dataCoordinatorV1.SearchReq{
		BucketName: "tbucket",
		Vector:     searchVec,
		TopK:       3,
		EfSearch:   16,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Status {
		for _, dd := range resp.GetCandidates() {
			meta := make(map[string]interface{})
			msgpack.Unmarshal(dd.GetMetadata(), &meta)
			fmt.Println(dd.Id, dd.Score, meta)
		}
	} else {
		log.Fatal(res.Error.ErrorMessage)
	}
	//add after data
	for _, data := range afterDataset {
		res, err := dclient.Insert(context.Background(), data)
		fmt.Println(res.Status, res.Error, err)
	}
	// second search
	fmt.Println("old+new data search")
	resp, err = dclient.Search(context.Background(), &dataCoordinatorV1.SearchReq{
		BucketName: "tbucket",
		Vector:     searchVec,
		TopK:       5,
		EfSearch:   16,
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Status {
		for _, dd := range resp.GetCandidates() {
			meta := make(map[string]interface{})
			msgpack.Unmarshal(dd.GetMetadata(), &meta)
			fmt.Println(dd.Id, dd.Score, meta)
		}
	} else {
		log.Fatal(res.Error.ErrorMessage)
	}
```

The current code is highly experimental and focuses solely on the initial phase of verifying whether search functionality works. This code is unstable and intended as proof-of-concept (PoC) code. It performs additions and displays individual search results within a similar dataset (only small-scale datasets have been tested so far). (Note: Search results are also currently unsorted.)

#### Result

```sh
<nil>
<nil> false
true <nil> <nil>
true <nil> <nil>
true <nil> <nil>
true <nil> <nil>
true <nil> <nil>
old data search
old data search
 15 map[_id:6a1f0c33-00ec-4054-9c95-126c2c3af548 description:I like fixing cars]
 15 map[_id:6a1f0c33-00ec-4054-9c95-126c2c3af548 description:I like fixing cars]
 50.3 map[_id:7737effd-7990-4be3-bcde-0a843f425c6c description:I usually eat a sandwich]
 50.3 map[_id:7737effd-7990-4be3-bcde-0a843f425c6c description:I usually eat a sandwich]
 64.6 map[_id:95b53c24-7811-4899-a374-3a71ba0e2243 description:When I'm hungry, I tend 64.6 map[_id:95b53c24-7811-4899-a374-3a71ba0e2243 description:When I'm hungry, I tend to eat porridge]
 to eat porridge]
true <nil> <nil>
true <nil> <nil>
 27.2 map[_id:280fe999-ba7e-40ea-9eac-c9d6ca2d0866 description:Someday, I'll become a great person]
 64.6 map[_id:95b53c24-7811-4899-a374-3a71ba0e2243 description:When I'm hungry, I tend to eat porridge]
 50.3 map[_id:7737effd-7990-4be3-bcde-0a843f425c6c description:I usually eat a sandwich]
 67.6 map[_id:b4d03634-9fdb-4416-b4be-0c31a9e4ea54 description:I've thought about what to eat when I'm hungry, and cookies are definitely the best]
```
