	.section	__TEXT,__text,regular,pure_instructions
	.build_version macos, 10, 15	sdk_version 10, 15
	.intel_syntax noprefix
	.globl	__Z26euclidean_distance_squaredmPfS_S_ ## -- Begin function _Z26euclidean_distance_squaredmPfS_S_
	.p2align	4, 0x90
__Z26euclidean_distance_squaredmPfS_S_: ## @_Z26euclidean_distance_squaredmPfS_S_
## %bb.0:
	push	rbp
	mov	rbp, rsp
	and	rsp, -8
	mov	r8, rdi
	and	r8, -4
	je	LBB0_1
## %bb.2:
	lea	r10, [r8 - 1]
	mov	rax, r10
	shr	rax, 2
	lea	r9d, [rax + 1]
	and	r9d, 3
	cmp	r10, 12
	jae	LBB0_12
## %bb.3:
	xorps	xmm0, xmm0
	xor	eax, eax
	test	r9, r9
	jne	LBB0_5
	jmp	LBB0_7
LBB0_1:
	xorps	xmm0, xmm0
	jmp	LBB0_7
LBB0_12:
	lea	r10, [r9 - 1]
	sub	r10, rax
	xorps	xmm1, xmm1
	xor	eax, eax
	.p2align	4, 0x90
LBB0_13:                                ## =>This Inner Loop Header: Depth=1
	movaps	xmm2, xmmword ptr [rsi + 4*rax]
	movaps	xmm3, xmmword ptr [rsi + 4*rax + 16]
	movaps	xmm4, xmmword ptr [rsi + 4*rax + 32]
	movaps	xmm0, xmmword ptr [rsi + 4*rax + 48]
	subps	xmm2, xmmword ptr [rdx + 4*rax]
	mulps	xmm2, xmm2
	addps	xmm2, xmm1
	subps	xmm3, xmmword ptr [rdx + 4*rax + 16]
	mulps	xmm3, xmm3
	addps	xmm3, xmm2
	subps	xmm4, xmmword ptr [rdx + 4*rax + 32]
	mulps	xmm4, xmm4
	addps	xmm4, xmm3
	subps	xmm0, xmmword ptr [rdx + 4*rax + 48]
	mulps	xmm0, xmm0
	addps	xmm0, xmm4
	add	rax, 16
	movaps	xmm1, xmm0
	add	r10, 4
	jne	LBB0_13
## %bb.4:
	test	r9, r9
	je	LBB0_7
LBB0_5:
	shl	rax, 2
	neg	r9
	.p2align	4, 0x90
LBB0_6:                                 ## =>This Inner Loop Header: Depth=1
	movaps	xmm1, xmmword ptr [rsi + rax]
	subps	xmm1, xmmword ptr [rdx + rax]
	mulps	xmm1, xmm1
	addps	xmm0, xmm1
	add	rax, 16
	inc	r9
	jne	LBB0_6
LBB0_7:
	movshdup	xmm1, xmm0      ## xmm1 = xmm0[1,1,3,3]
	addss	xmm1, xmm0
	movaps	xmm2, xmm0
	unpckhpd	xmm2, xmm0      ## xmm2 = xmm2[1],xmm0[1]
	addss	xmm2, xmm1
	shufps	xmm0, xmm0, 231         ## xmm0 = xmm0[3,1,2,3]
	addss	xmm0, xmm2
	movss	dword ptr [rcx], xmm0
	movsxd	rax, r8d
	cmp	rax, rdi
	jae	LBB0_11
## %bb.8:
	movsxd	rax, r8d
	movsxd	r8, edi
	not	r8
	or	r8, 3
	test	dil, 1
	je	LBB0_10
## %bb.9:
	movss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	subss	xmm1, dword ptr [rdx + 4*rax]
	mulss	xmm1, xmm1
	addss	xmm0, xmm1
	movss	dword ptr [rcx], xmm0
	or	rax, 1
LBB0_10:
	add	r8, rdi
	je	LBB0_11
	.p2align	4, 0x90
LBB0_14:                                ## =>This Inner Loop Header: Depth=1
	movss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	subss	xmm1, dword ptr [rdx + 4*rax]
	mulss	xmm1, xmm1
	addss	xmm1, xmm0
	movss	dword ptr [rcx], xmm1
	movss	xmm0, dword ptr [rsi + 4*rax + 4] ## xmm0 = mem[0],zero,zero,zero
	subss	xmm0, dword ptr [rdx + 4*rax + 4]
	mulss	xmm0, xmm0
	addss	xmm0, xmm1
	movss	dword ptr [rcx], xmm0
	add	rax, 2
	cmp	rdi, rax
	jne	LBB0_14
LBB0_11:
	mov	rsp, rbp
	pop	rbp
	ret
                                        ## -- End function
	.section	__TEXT,__literal16,16byte_literals
	.p2align	4               ## -- Begin function _Z18manhattan_distancemPfS_S_
LCPI1_0:
	.long	2147483648              ## float -0
	.long	2147483648              ## float -0
	.long	2147483648              ## float -0
	.long	2147483648              ## float -0
	.section	__TEXT,__text,regular,pure_instructions
	.globl	__Z18manhattan_distancemPfS_S_
	.p2align	4, 0x90
__Z18manhattan_distancemPfS_S_:         ## @_Z18manhattan_distancemPfS_S_
## %bb.0:
	push	rbp
	mov	rbp, rsp
	and	rsp, -8
	mov	r9, rdi
	and	r9, -4
	je	LBB1_1
## %bb.2:
	lea	rax, [r9 - 1]
	shr	rax, 2
	lea	r8d, [rax + 1]
	and	r8d, 1
	test	rax, rax
	je	LBB1_3
## %bb.13:
	lea	r10, [r8 - 1]
	sub	r10, rax
	xorps	xmm0, xmm0
	xor	eax, eax
	.p2align	4, 0x90
LBB1_14:                                ## =>This Inner Loop Header: Depth=1
	movaps	xmm1, xmmword ptr [rsi + 4*rax]
	movaps	xmm2, xmmword ptr [rsi + 4*rax + 16]
	subps	xmm1, xmmword ptr [rdx + 4*rax]
	mulps	xmm1, xmm1
	sqrtps	xmm1, xmm1
	addps	xmm1, xmm0
	subps	xmm2, xmmword ptr [rdx + 4*rax + 16]
	mulps	xmm2, xmm2
	sqrtps	xmm0, xmm2
	addps	xmm0, xmm1
	add	rax, 8
	add	r10, 2
	jne	LBB1_14
## %bb.4:
	test	r8, r8
	jne	LBB1_5
	jmp	LBB1_6
LBB1_1:
	xorps	xmm0, xmm0
	jmp	LBB1_6
LBB1_3:
	xorps	xmm0, xmm0
	xor	eax, eax
	test	r8, r8
	je	LBB1_6
LBB1_5:
	movaps	xmm1, xmmword ptr [rsi + 4*rax]
	subps	xmm1, xmmword ptr [rdx + 4*rax]
	mulps	xmm1, xmm1
	sqrtps	xmm1, xmm1
	addps	xmm0, xmm1
LBB1_6:
	movshdup	xmm1, xmm0      ## xmm1 = xmm0[1,1,3,3]
	addss	xmm1, xmm0
	movaps	xmm2, xmm0
	unpckhpd	xmm2, xmm0      ## xmm2 = xmm2[1],xmm0[1]
	addss	xmm2, xmm1
	shufps	xmm0, xmm0, 231         ## xmm0 = xmm0[3,1,2,3]
	addss	xmm0, xmm2
	movss	dword ptr [rcx], xmm0
	movsxd	rax, r9d
	cmp	rax, rdi
	jae	LBB1_12
## %bb.7:
	movsxd	rax, r9d
	movsxd	r8, edi
	not	r8
	or	r8, 3
	test	dil, 1
	je	LBB1_9
## %bb.8:
	movss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	subss	xmm1, dword ptr [rdx + 4*rax]
	movaps	xmm2, xmmword ptr [rip + LCPI1_0] ## xmm2 = [-0.0E+0,-0.0E+0,-0.0E+0,-0.0E+0]
	xorps	xmm2, xmm1
	xorps	xmm3, xmm3
	movaps	xmm4, xmm1
	cmpltss	xmm4, xmm3
	movaps	xmm3, xmm4
	andnps	xmm3, xmm1
	andps	xmm4, xmm2
	orps	xmm4, xmm3
	addss	xmm0, xmm4
	movss	dword ptr [rcx], xmm0
	or	rax, 1
LBB1_9:
	add	r8, rdi
	je	LBB1_12
## %bb.10:
	movaps	xmm1, xmmword ptr [rip + LCPI1_0] ## xmm1 = [-0.0E+0,-0.0E+0,-0.0E+0,-0.0E+0]
	xorps	xmm2, xmm2
	.p2align	4, 0x90
LBB1_11:                                ## =>This Inner Loop Header: Depth=1
	movss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	subss	xmm3, dword ptr [rdx + 4*rax]
	movaps	xmm4, xmm3
	cmpltss	xmm4, xmm2
	movaps	xmm5, xmm4
	andnps	xmm5, xmm3
	xorps	xmm3, xmm1
	andps	xmm4, xmm3
	orps	xmm4, xmm5
	addss	xmm4, xmm0
	movss	dword ptr [rcx], xmm4
	movss	xmm3, dword ptr [rsi + 4*rax + 4] ## xmm3 = mem[0],zero,zero,zero
	subss	xmm3, dword ptr [rdx + 4*rax + 4]
	movaps	xmm0, xmm3
	cmpltss	xmm0, xmm2
	movaps	xmm5, xmm0
	andnps	xmm5, xmm3
	xorps	xmm3, xmm1
	andps	xmm0, xmm3
	orps	xmm0, xmm5
	addss	xmm0, xmm4
	movss	dword ptr [rcx], xmm0
	add	rax, 2
	cmp	rdi, rax
	jne	LBB1_11
LBB1_12:
	mov	rsp, rbp
	pop	rbp
	ret
                                        ## -- End function
	.globl	__Z26cosine_similarity_dot_normmPfS_S_S_ ## -- Begin function _Z26cosine_similarity_dot_normmPfS_S_S_
	.p2align	4, 0x90
__Z26cosine_similarity_dot_normmPfS_S_S_: ## @_Z26cosine_similarity_dot_normmPfS_S_S_
## %bb.0:
	push	rbp
	mov	rbp, rsp
	and	rsp, -8
	mov	r10, rdi
	and	r10, -4
	je	LBB2_1
## %bb.2:
	lea	rax, [r10 - 1]
	shr	rax, 2
	lea	r9d, [rax + 1]
	and	r9d, 1
	test	rax, rax
	je	LBB2_3
## %bb.11:
	lea	r11, [r9 - 1]
	sub	r11, rax
	xorps	xmm3, xmm3
	xor	eax, eax
	xorps	xmm4, xmm4
	xorps	xmm1, xmm1
	.p2align	4, 0x90
LBB2_12:                                ## =>This Inner Loop Header: Depth=1
	movaps	xmm5, xmmword ptr [rsi + 4*rax]
	movaps	xmm2, xmmword ptr [rsi + 4*rax + 16]
	movaps	xmm6, xmmword ptr [rdx + 4*rax]
	movaps	xmm0, xmmword ptr [rdx + 4*rax + 16]
	movaps	xmm7, xmm5
	mulps	xmm7, xmm6
	addps	xmm7, xmm1
	mulps	xmm5, xmm5
	addps	xmm5, xmm4
	mulps	xmm6, xmm6
	addps	xmm6, xmm3
	movaps	xmm1, xmm2
	mulps	xmm1, xmm0
	addps	xmm1, xmm7
	mulps	xmm2, xmm2
	addps	xmm2, xmm5
	mulps	xmm0, xmm0
	addps	xmm0, xmm6
	add	rax, 8
	movaps	xmm3, xmm0
	movaps	xmm4, xmm2
	add	r11, 2
	jne	LBB2_12
## %bb.4:
	test	r9, r9
	jne	LBB2_5
	jmp	LBB2_6
LBB2_1:
	xorps	xmm1, xmm1
	xorps	xmm2, xmm2
	xorps	xmm0, xmm0
	jmp	LBB2_6
LBB2_3:
	xorps	xmm0, xmm0
	xor	eax, eax
	xorps	xmm2, xmm2
	xorps	xmm1, xmm1
	test	r9, r9
	je	LBB2_6
LBB2_5:
	movaps	xmm3, xmmword ptr [rsi + 4*rax]
	movaps	xmm4, xmmword ptr [rdx + 4*rax]
	movaps	xmm5, xmm3
	mulps	xmm3, xmm4
	mulps	xmm4, xmm4
	addps	xmm0, xmm4
	mulps	xmm5, xmm5
	addps	xmm2, xmm5
	addps	xmm1, xmm3
LBB2_6:
	movshdup	xmm3, xmm1      ## xmm3 = xmm1[1,1,3,3]
	addss	xmm3, xmm1
	movaps	xmm4, xmm1
	unpckhpd	xmm4, xmm1      ## xmm4 = xmm4[1],xmm1[1]
	addss	xmm4, xmm3
	shufps	xmm1, xmm1, 231         ## xmm1 = xmm1[3,1,2,3]
	addss	xmm1, xmm4
	movss	dword ptr [rcx], xmm1
	movaps	xmm3, xmm2
	insertps	xmm3, xmm0, 28  ## xmm3 = xmm3[0],xmm0[0],zero,zero
	movaps	xmm4, xmm0
	insertps	xmm4, xmm2, 76  ## xmm4 = xmm2[1],xmm4[1],zero,zero
	addps	xmm4, xmm3
	movaps	xmm3, xmm2
	unpckhpd	xmm3, xmm2      ## xmm3 = xmm3[1],xmm2[1]
	insertps	xmm3, xmm0, 156 ## xmm3 = xmm3[0],xmm0[2],zero,zero
	addps	xmm3, xmm4
	movhlps	xmm0, xmm0              ## xmm0 = xmm0[1,1]
	insertps	xmm0, xmm2, 204 ## xmm0 = xmm2[3],xmm0[1],zero,zero
	addps	xmm0, xmm3
	movsxd	rax, r10d
	cmp	rax, rdi
	jae	LBB2_10
## %bb.7:
	movsxd	rax, r10d
	movsxd	r9, edi
	not	r9
	or	r9, 3
	test	dil, 1
	je	LBB2_9
## %bb.8:
	movss	xmm2, dword ptr [rsi + 4*rax] ## xmm2 = mem[0],zero,zero,zero
	mulss	xmm2, dword ptr [rdx + 4*rax]
	addss	xmm1, xmm2
	movss	dword ptr [rcx], xmm1
	movss	xmm2, dword ptr [rsi + 4*rax] ## xmm2 = mem[0],zero,zero,zero
	insertps	xmm2, dword ptr [rdx + 4*rax], 16 ## xmm2 = xmm2[0],mem[0],xmm2[2,3]
	mulps	xmm2, xmm2
	addps	xmm0, xmm2
	or	rax, 1
LBB2_9:
	add	r9, rdi
	je	LBB2_10
	.p2align	4, 0x90
LBB2_13:                                ## =>This Inner Loop Header: Depth=1
	movss	xmm2, dword ptr [rsi + 4*rax] ## xmm2 = mem[0],zero,zero,zero
	mulss	xmm2, dword ptr [rdx + 4*rax]
	addss	xmm2, xmm1
	movss	dword ptr [rcx], xmm2
	movss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	movss	xmm1, dword ptr [rsi + 4*rax + 4] ## xmm1 = mem[0],zero,zero,zero
	insertps	xmm3, dword ptr [rdx + 4*rax], 16 ## xmm3 = xmm3[0],mem[0],xmm3[2,3]
	mulps	xmm3, xmm3
	addps	xmm3, xmm0
	mulss	xmm1, dword ptr [rdx + 4*rax + 4]
	addss	xmm1, xmm2
	movss	dword ptr [rcx], xmm1
	movss	xmm0, dword ptr [rsi + 4*rax + 4] ## xmm0 = mem[0],zero,zero,zero
	insertps	xmm0, dword ptr [rdx + 4*rax + 4], 16 ## xmm0 = xmm0[0],mem[0],xmm0[2,3]
	mulps	xmm0, xmm0
	addps	xmm0, xmm3
	add	rax, 2
	cmp	rdi, rax
	jne	LBB2_13
LBB2_10:
	movshdup	xmm1, xmm0      ## xmm1 = xmm0[1,1,3,3]
	mulss	xmm1, xmm0
	movss	dword ptr [r8], xmm1
	mov	rsp, rbp
	pop	rbp
	ret
                                        ## -- End function

.subsections_via_symbols