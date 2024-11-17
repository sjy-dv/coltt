from sentence_transformers import SentenceTransformer
import pandas as pd
import json
import numpy as np
model = SentenceTransformer('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2')


data = []

q1 = ['Where is a place with a comfortable indoor environment?']
q2 = ['Where is a good restaurant to go to with family?']
q3 = ['A restaurant I strongly do not recommend.']
# data.append({'review': q1, 'embedding':model.encode(q1).tolist()})
# data.append({'review': q2, 'embedding':model.encode(q2).tolist()})
# data.append({'review': q3, 'embedding':model.encode(q3).tolist()})

# json_file_path = 'review_question.json'
# with open(json_file_path, 'w', encoding='utf-8') as f:
#     json.dump(data, f, ensure_ascii=False, indent=4)

#  normalize section

embeddings1q = model.encode(q1)
embeddings1q = embeddings1q / np.linalg.norm(embeddings1q, axis=1, keepdims=True)
embeddings2q = model.encode(q1)
embeddings2q = embeddings2q / np.linalg.norm(embeddings2q, axis=1, keepdims=True)
embeddings3q = model.encode(q1)
embeddings3q = embeddings3q / np.linalg.norm(embeddings3q, axis=1, keepdims=True)

data.append({'review': q1[0], 'embedding':embeddings1q[0].tolist()})
data.append({'review': q2[0], 'embedding':embeddings2q[0].tolist()})
data.append({'review': q3[0], 'embedding':embeddings3q[0].tolist()})

json_file_path = 'normalize_review_question.json'
with open(json_file_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)