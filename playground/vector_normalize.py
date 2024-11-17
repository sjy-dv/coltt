from sentence_transformers import SentenceTransformer
import numpy as np
import json

model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')

sentences = ["This is a good temperature for development."]

embd = model.encode(sentences[0]).tolist()
embeddings = model.encode(sentences)
# save to using go
data = []
data.append({'embedding':embd})

with open("normalize_test.json", 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)



embeddings = embeddings / np.linalg.norm(embeddings, axis=1, keepdims=True)

print(embeddings.shape)
points = []


ndata = []
ndata.append({'embedding': embeddings[0].tolist()})

with open("normalize_output_py.json", 'w', encoding='utf-8') as f:
    json.dump(ndata, f, ensure_ascii=False, indent=4)