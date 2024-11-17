from sentence_transformers import SentenceTransformer
import pandas as pd
import json

# Load the model
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

# Load the CSV file
# csv_file_path = 'dataset.csv'
csv_file_path = 'reviews.csv'
df = pd.read_csv(csv_file_path)

# Process the data and generate embeddings
data = []
for review in df['text']:
    embedding = model.encode(review).tolist()  # Convert to list to make JSON serializable
    data.append({'review': review, 'embedding': embedding})

# Save to JSON file
json_file_path = 'short_text.json'
with open(json_file_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print(f"Data has been written to {json_file_path}")

# normalize section

# from sentence_transformers import SentenceTransformer
# import pandas as pd
# import json
# import numpy as np

# # Load the model
# model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

# # Load the CSV file
# # csv_file_path = 'dataset.csv'
# csv_file_path = 'reviews.csv'
# df = pd.read_csv(csv_file_path)

# # Process the data and generate embeddings
# data = []
# for review in df['text']:
#     embedding = model.encode([review])  # Convert to list to make JSON serializable
#     embedding = embedding / np.linalg.norm(embedding, axis=1, keepdims=True)
#     data.append({'review': review, 'embedding': embedding[0].tolist()})

# # Save to JSON file
# json_file_path = 'normalize_short_text.json'
# with open(json_file_path, 'w', encoding='utf-8') as f:
#     json.dump(data, f, ensure_ascii=False, indent=4)

# print(f"Data has been written to {json_file_path}")
