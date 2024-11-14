import chromadb
import pandas as pd
import json
import uuid  # For generating unique IDs
import logging
import time  # For measuring latency

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize ChromaDB client
client = chromadb.Client()

# Create or get the collection
collection_name = 'compare_review_collection'
if collection_name in client.list_collections():
    collection = client.get_collection(collection_name)
    logger.info(f"Collection '{collection_name}' found and loaded.")
else:
    collection = client.create_collection(name=collection_name, metadata={'hnsw:space':'l2'})
    logger.info(f"Collection '{collection_name}' created.")

# Path to your JSON file containing short texts
json_path = "short_text.json"

# Read the JSON file into a pandas DataFrame
try:
    df = pd.read_json(json_path)
    logger.info(f"Loaded {len(df)} records from {json_path}.")
except ValueError as e:
    logger.error(f"Error reading JSON file at {json_path}: {e}")
    exit(1)

# Check if required columns exist
required_columns = {'embedding', 'review'}
if not required_columns.issubset(df.columns):
    logger.error(f"JSON file must contain the following columns: {required_columns}")
    exit(1)

# Define expected embedding dimensions (adjust as needed)
expected_dim = None  # Set to a specific integer if known

# Add embeddings and metadata to the collection
for index, row in df.iterrows():
    embedding = row['embedding']
    review = row['review']
    
    # Validate the embedding
    if not isinstance(embedding, (list, tuple)):
        logger.warning(f"Invalid embedding type at row {index}: Expected list or tuple, got {type(embedding)}")
        continue  # Skip this row
    
    if expected_dim and len(embedding) != expected_dim:
        logger.warning(f"Invalid embedding dimension at row {index}: Expected {expected_dim}, got {len(embedding)}")
        continue  # Skip this row
    
    # Generate a unique ID for each item
    unique_id = str(uuid.uuid4())
    
    # Add to collection
    try:
        collection.add(
            ids=[unique_id],
            embeddings=[embedding],
            metadatas=[{"review": review}]
        )
        logger.info(f"Added row {index} with ID {unique_id} to the collection.")
    except Exception as e:
        logger.error(f"Error adding row {index} to collection: {e}")
        continue

# Path to your JSON file containing review questions
qpath = "review_question.json"

# Read the query JSON file into a pandas DataFrame
try:
    qf = pd.read_json(qpath)
    logger.info(f"Loaded {len(qf)} query records from {qpath}.")
except ValueError as e:
    logger.error(f"Error reading JSON file at {qpath}: {e}")
    exit(1)

# Check if required columns exist in query DataFrame
required_query_columns = {'embedding', 'review'}
if not required_query_columns.issubset(qf.columns):
    logger.error(f"Query JSON file must contain the following columns: {required_query_columns}")
    exit(1)

# Initialize a list to hold all ResultCompare entries
result_compare_list = []

# Perform queries
for index, q in qf.iterrows():
    query_embedding = q['embedding']
    query_review = q['review']
    
    # Validate the query embedding
    if not isinstance(query_embedding, (list, tuple)):
        logger.warning(f"Invalid query embedding type at row {index}: Expected list or tuple, got {type(query_embedding)}")
        continue  # Skip this query
    
    if expected_dim and len(query_embedding) != expected_dim:
        logger.warning(f"Invalid query embedding dimension at row {index}: Expected {expected_dim}, got {len(query_embedding)}")
        continue  # Skip this query
    
    logger.info(f"Processing Query {index + 1}: {query_review}")
    
    # Measure the latency
    start_time = time.time()
    try:
        results = collection.query(query_embeddings=[query_embedding], n_results=5)
        end_time = time.time()
        latency = f"{(end_time - start_time) * 1000:.2f} ms"  # Convert to milliseconds
    except Exception as e:
        logger.error(f"Error querying with row {index}: {e}")
        continue
    
    # Extract similar reviews
    similar_reviews = []
    # The structure of 'results' may vary based on ChromaDB's API version
    # Adjust the following lines accordingly
    # Assuming 'metadatas' is a list of lists (one per query)
    try:
        metadatas = results.get('metadatas', [])
        if metadatas and isinstance(metadatas, list):
            for metadata in metadatas[0]:  # Since we have one query
                review_text = metadata.get('review', '')
                if review_text:
                    similar_reviews.append({"review": review_text})
    except Exception as e:
        logger.error(f"Error extracting metadatas for query {index}: {e}")
    
    # Create a ResultCompare entry
    result_compare = {
        "base_review": query_review,
        "similar_review": similar_reviews,
        "latency": latency
    }
    
    # Append to the list
    result_compare_list.append(result_compare)
    
    # Optionally, print the result
    logger.info(f"Results for Query {index + 1}: {similar_reviews}")
    logger.info(f"Latency: {latency}")

# Save all results to a JSON file
output_path = "compare_chroma_short_text_results.json"
try:
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(result_compare_list, f, ensure_ascii=False, indent=4)
    logger.info(f"All results have been saved to {output_path}.")
except Exception as e:
    logger.error(f"Error saving results to {output_path}: {e}")
