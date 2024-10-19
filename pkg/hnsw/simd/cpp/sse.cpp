#include <immintrin.h>

inline void _sum_vector(__m128 vec, float *result)
{
    *result = (((float *)&vec)[0] + ((float *)&vec)[1] + ((float *)&vec)[2] + ((float *)&vec)[3]);
}

inline float abs(float x)
{
    return (x < 0) ? -x : x;
}

void euclidean_distance_squared(size_t len, float *a, float *b, float *result)
{
    __m128 v1, v2;
    __m128 resVec = _mm_setzero_ps();
    for (int i = 0; i < (len / 4) * 4; i += 4)
    {
        v1 = _mm_load_ps(&a[i]);
        v2 = _mm_load_ps(&b[i]);
        v1 = _mm_sub_ps(v1, v2);
        v2 = _mm_mul_ps(v1, v1);
        resVec = _mm_add_ps(resVec, v2);
    }
    _sum_vector(resVec, result);

    float diff;
    for (int i = (len / 4) * 4; i < len; i++)
    {
        diff = a[i] - b[i];
        *result += diff * diff;
    }
}

void manhattan_distance(size_t len, float *a, float *b, float *result)
{
    __m128 v1, v2;
    __m128 resVec = _mm_setzero_ps();
    for (int i = 0; i < (len / 4) * 4; i += 4)
    {
        v1 = _mm_load_ps(&a[i]);
        v2 = _mm_load_ps(&b[i]);
        v1 = _mm_sub_ps(v1, v2);
        v2 = _mm_sqrt_ps(_mm_mul_ps(v1, v1));
        resVec = _mm_add_ps(resVec, v2);
    }
    _sum_vector(resVec, result);

    for (int i = (len / 4) * 4; i < len; i++)
    {
        *result += abs(a[i] - b[i]);
    }
}

void cosine_similarity_dot_norm(size_t len, float *a, float *b, float *result_dot, float *result_norm_squared)
{
    __m128 v1, v2, dot, norm_a, norm_b;
    float norm_a_sum, norm_b_sum;
    dot = _mm_setzero_ps();
    norm_a = _mm_setzero_ps();
    norm_b = _mm_setzero_ps();
    for (int i = 0; i < (len / 4) * 4; i += 4)
    {
        v1 = _mm_load_ps(&a[i]);
        v2 = _mm_load_ps(&b[i]);
        dot = _mm_add_ps(dot, _mm_mul_ps(v1, v2));
        norm_a = _mm_add_ps(norm_a, _mm_mul_ps(v1, v1));
        norm_b = _mm_add_ps(norm_b, _mm_mul_ps(v2, v2));
    }

    _sum_vector(dot, result_dot);
    _sum_vector(norm_a, &norm_a_sum);
    _sum_vector(norm_b, &norm_b_sum);

    for (int i = (len / 4) * 4; i < len; i++)
    {
        *result_dot += a[i] * b[i];
        norm_a_sum += a[i] * a[i];
        norm_b_sum += b[i] * b[i];
    }

    *result_norm_squared = (norm_a_sum * norm_b_sum);
}