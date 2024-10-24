#include <immintrin.h>

inline
void _sum_vector(__m256 vec, float *result) {
	vec = _mm256_hadd_ps(vec, vec);
	vec = _mm256_hadd_ps(vec, vec);
	*result = (((float*)&vec)[0] + ((float*)&vec)[4]);	
}

inline
float abs(float x) {
	return (x < 0) ? -x : x;
}

void euclidean_distance_squared(size_t len, float *a, float *b, float *result) {
	__m256 v1, v2;
	__m256 resVec = _mm256_setzero_ps();
	for (int i = 0; i < (len / 8) * 8; i += 8) {
		v1 = _mm256_load_ps(&a[i]);
		v2 = _mm256_load_ps(&b[i]);
		v1 = _mm256_sub_ps(v1, v2);
		v2 = _mm256_mul_ps(v1, v1);
		resVec = _mm256_add_ps(resVec, v2);
	}
	_sum_vector(resVec, result);

	float diff;
	for (int i = (len / 8) * 8; i < len; i++) {
		diff = a[i] - b[i];
		*result += diff * diff;
	}
}

void manhattan_distance(size_t len, float *a, float *b, float *result) {
	__m256 v1, v2;
	__m256 resVec = _mm256_setzero_ps();
	for (int i = 0; i < (len / 8) * 8; i += 8) {
		v1 = _mm256_load_ps(&a[i]);
		v2 = _mm256_load_ps(&b[i]);
		v1 = _mm256_sub_ps(v1, v2);
		v2 = _mm256_sqrt_ps(_mm256_mul_ps(v1, v1));
		resVec = _mm256_add_ps(resVec, v2);
	}
	_sum_vector(resVec, result);

	for (int i = (len / 8) * 8; i < len; i++) {
		*result += abs(a[i] - b[i]);
	}
}

void cosine_similarity_dot_norm(size_t len, float *a, float *b, float *result_dot, float *result_norm_squared) {
	__m256 v1, v2, dot, norm_a, norm_b;
	float norm_a_sum, norm_b_sum;
	dot = _mm256_setzero_ps();
	norm_a = _mm256_setzero_ps();
	norm_b = _mm256_setzero_ps();
	for (int i = 0; i < (len / 8) * 8; i += 8) {
		v1 = _mm256_load_ps(&a[i]);
		v2 = _mm256_load_ps(&b[i]);
		dot = _mm256_add_ps(dot, _mm256_mul_ps(v1, v2));
		norm_a = _mm256_add_ps(norm_a, _mm256_mul_ps(v1, v1));
		norm_b = _mm256_add_ps(norm_b, _mm256_mul_ps(v2, v2));
	}

	_sum_vector(dot, result_dot);
	_sum_vector(norm_a, &norm_a_sum);
	_sum_vector(norm_b, &norm_b_sum);

	for (int i = (len / 8) * 8; i < len; i++) {
		*result_dot += a[i] * b[i];
		norm_a_sum += a[i] * a[i];
		norm_b_sum += b[i] * b[i];
	}

	*result_norm_squared = (norm_a_sum * norm_b_sum);
}