### I will explain the internal data storage flow.

![arch6](./assets/arch6.png)

First, HNSW operates in memory internally, and its data is stored as cached files. However, this poses a risk of data corruption in the event of an abnormal shutdown.

To address this, nnlogdb (no-named-tsdb) is internally deployed to track insert, update, and delete events. Since only metadata and vectors are needed without node links, this is not a significant issue.

The observer continuously compares the tracked log values with the latest node, and if a problem arises, HNSW recovery is initiated.

### Disk files can sometimes become corrupted and fail to open, leading to significant issues. Is cached data safe?

![arch7](./assets/arch7.png)
Cache data files support fast loading and saving through the INFLATE/DEFLATE compression algorithm. However, cache files are inherently much less stable than disk files.

To address this, we deploy "old" versions. These versions are not user-specified; instead, they are managed internally. During idle periods, data changes are saved as new cache data, while the previous stable open file version is stored as the "old" version. When this happens, the last update time of the "old" version aligns with the sync time in nnlogdb.

To manage disk usage efficiently, all previous partitions up to the reliably synced period are deleted.

This approach ensures stable data management.

### Does disk storage also have structural flexibility?

Disk storage will not initially have structural flexibility. However, in the long term, we aim to either introduce flexibility for the disk structure or, unfortunately, impose some restrictions on the memory side. While no final decision has been made, we believe memory storage should maintain flexibility, so we’re likely to design disk storage with some degree of structural flexibility in the future. SQLite will be supported as the disk storage option.
