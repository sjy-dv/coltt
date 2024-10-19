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
	and	r8, -8
	je	LBB0_1
## %bb.2:
	lea	r10, [r8 - 1]
	mov	rax, r10
	shr	rax, 3
	lea	r9d, [rax + 1]
	and	r9d, 3
	cmp	r10, 24
	jae	LBB0_12
## %bb.3:
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	test	r9, r9
	jne	LBB0_5
	jmp	LBB0_7
LBB0_1:
	vxorps	xmm0, xmm0, xmm0
	jmp	LBB0_7
LBB0_12:
	lea	r10, [r9 - 1]
	sub	r10, rax
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	.p2align	4, 0x90
LBB0_13:                                ## =>This Inner Loop Header: Depth=1
	vmovups	ymm1, ymmword ptr [rsi + 4*rax]
	vmovups	ymm2, ymmword ptr [rsi + 4*rax + 32]
	vmovups	ymm3, ymmword ptr [rsi + 4*rax + 64]
	vmovups	ymm4, ymmword ptr [rsi + 4*rax + 96]
	vsubps	ymm1, ymm1, ymmword ptr [rdx + 4*rax]
	vmulps	ymm1, ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
	vsubps	ymm1, ymm2, ymmword ptr [rdx + 4*rax + 32]
	vmulps	ymm1, ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
	vsubps	ymm1, ymm3, ymmword ptr [rdx + 4*rax + 64]
	vmulps	ymm1, ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
	vsubps	ymm1, ymm4, ymmword ptr [rdx + 4*rax + 96]
	vmulps	ymm1, ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
	add	rax, 32
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
	vmovups	ymm1, ymmword ptr [rsi + rax]
	vsubps	ymm1, ymm1, ymmword ptr [rdx + rax]
	vmulps	ymm1, ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
	add	rax, 32
	inc	r9
	jne	LBB0_6
LBB0_7:
	vhaddps	ymm0, ymm0, ymm0
	vhaddps	ymm0, ymm0, ymm0
	vextractf128	xmm1, ymm0, 1
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	movsxd	rax, r8d
	cmp	rax, rdi
	jae	LBB0_11
## %bb.8:
	movsxd	rax, r8d
	movsxd	r8, edi
	not	r8
	or	r8, 7
	test	dil, 1
	je	LBB0_10
## %bb.9:
	vmovss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	vsubss	xmm1, xmm1, dword ptr [rdx + 4*rax]
	vmulss	xmm1, xmm1, xmm1
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	or	rax, 1
LBB0_10:
	add	r8, rdi
	je	LBB0_11
	.p2align	4, 0x90
LBB0_14:                                ## =>This Inner Loop Header: Depth=1
	vmovss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	vsubss	xmm1, xmm1, dword ptr [rdx + 4*rax]
	vmulss	xmm1, xmm1, xmm1
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	vmovss	xmm1, dword ptr [rsi + 4*rax + 4] ## xmm1 = mem[0],zero,zero,zero
	vsubss	xmm1, xmm1, dword ptr [rdx + 4*rax + 4]
	vmulss	xmm1, xmm1, xmm1
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	add	rax, 2
	cmp	rdi, rax
	jne	LBB0_14
LBB0_11:
	mov	rsp, rbp
	pop	rbp
	vzeroupper
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
	and	r9, -8
	je	LBB1_1
## %bb.2:
	lea	rax, [r9 - 1]
	shr	rax, 3
	lea	r8d, [rax + 1]
	and	r8d, 1
	test	rax, rax
	je	LBB1_3
## %bb.13:
	lea	r10, [r8 - 1]
	sub	r10, rax
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	.p2align	4, 0x90
LBB1_14:                                ## =>This Inner Loop Header: Depth=1
	vmovups	ymm1, ymmword ptr [rsi + 4*rax]
	vsubps	ymm1, ymm1, ymmword ptr [rdx + 4*rax]
	vmulps	ymm1, ymm1, ymm1
	vsqrtps	ymm1, ymm1
	vmovups	ymm2, ymmword ptr [rsi + 4*rax + 32]
	vsubps	ymm2, ymm2, ymmword ptr [rdx + 4*rax + 32]
	vmulps	ymm2, ymm2, ymm2
	vsqrtps	ymm2, ymm2
	vaddps	ymm0, ymm0, ymm1
	vaddps	ymm0, ymm0, ymm2
	add	rax, 16
	add	r10, 2
	jne	LBB1_14
## %bb.4:
	test	r8, r8
	jne	LBB1_5
	jmp	LBB1_6
LBB1_1:
	vxorps	xmm0, xmm0, xmm0
	jmp	LBB1_6
LBB1_3:
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	test	r8, r8
	je	LBB1_6
LBB1_5:
	vmovups	ymm1, ymmword ptr [rsi + 4*rax]
	vsubps	ymm1, ymm1, ymmword ptr [rdx + 4*rax]
	vmulps	ymm1, ymm1, ymm1
	vsqrtps	ymm1, ymm1
	vaddps	ymm0, ymm0, ymm1
LBB1_6:
	vhaddps	ymm0, ymm0, ymm0
	vhaddps	ymm0, ymm0, ymm0
	vextractf128	xmm1, ymm0, 1
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	movsxd	rax, r9d
	cmp	rax, rdi
	jae	LBB1_12
## %bb.7:
	movsxd	rax, r9d
	movsxd	r8, edi
	not	r8
	or	r8, 7
	test	dil, 1
	je	LBB1_9
## %bb.8:
	vmovss	xmm1, dword ptr [rsi + 4*rax] ## xmm1 = mem[0],zero,zero,zero
	vsubss	xmm1, xmm1, dword ptr [rdx + 4*rax]
	vxorps	xmm2, xmm1, xmmword ptr [rip + LCPI1_0]
	vxorps	xmm3, xmm3, xmm3
	vcmpltss	xmm3, xmm1, xmm3
	vblendvps	xmm1, xmm1, xmm2, xmm3
	vaddss	xmm0, xmm0, xmm1
	vmovss	dword ptr [rcx], xmm0
	or	rax, 1
LBB1_9:
	add	r8, rdi
	je	LBB1_12
## %bb.10:
	vmovaps	xmm1, xmmword ptr [rip + LCPI1_0] ## xmm1 = [-0.0E+0,-0.0E+0,-0.0E+0,-0.0E+0]
	vxorps	xmm2, xmm2, xmm2
	.p2align	4, 0x90
LBB1_11:                                ## =>This Inner Loop Header: Depth=1
	vmovss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vsubss	xmm3, xmm3, dword ptr [rdx + 4*rax]
	vxorps	xmm4, xmm3, xmm1
	vcmpltss	xmm5, xmm3, xmm2
	vblendvps	xmm3, xmm3, xmm4, xmm5
	vaddss	xmm0, xmm0, xmm3
	vmovss	dword ptr [rcx], xmm0
	vmovss	xmm3, dword ptr [rsi + 4*rax + 4] ## xmm3 = mem[0],zero,zero,zero
	vsubss	xmm3, xmm3, dword ptr [rdx + 4*rax + 4]
	vxorps	xmm4, xmm3, xmm1
	vcmpltss	xmm5, xmm3, xmm2
	vblendvps	xmm3, xmm3, xmm4, xmm5
	vaddss	xmm0, xmm0, xmm3
	vmovss	dword ptr [rcx], xmm0
	add	rax, 2
	cmp	rdi, rax
	jne	LBB1_11
LBB1_12:
	mov	rsp, rbp
	pop	rbp
	vzeroupper
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
	and	r10, -8
	je	LBB2_1
## %bb.2:
	lea	rax, [r10 - 1]
	shr	rax, 3
	lea	r9d, [rax + 1]
	and	r9d, 1
	test	rax, rax
	je	LBB2_3
## %bb.11:
	lea	r11, [r9 - 1]
	sub	r11, rax
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	vxorps	xmm2, xmm2, xmm2
	vxorps	xmm1, xmm1, xmm1
	.p2align	4, 0x90
LBB2_12:                                ## =>This Inner Loop Header: Depth=1
	vmovups	ymm3, ymmword ptr [rsi + 4*rax]
	vmovups	ymm4, ymmword ptr [rsi + 4*rax + 32]
	vmovups	ymm5, ymmword ptr [rdx + 4*rax]
	vmovups	ymm6, ymmword ptr [rdx + 4*rax + 32]
	vmulps	ymm7, ymm3, ymm5
	vaddps	ymm1, ymm1, ymm7
	vmulps	ymm3, ymm3, ymm3
	vaddps	ymm2, ymm2, ymm3
	vmulps	ymm3, ymm5, ymm5
	vaddps	ymm0, ymm0, ymm3
	vmulps	ymm3, ymm4, ymm6
	vaddps	ymm1, ymm1, ymm3
	vmulps	ymm3, ymm4, ymm4
	vaddps	ymm2, ymm2, ymm3
	vmulps	ymm3, ymm6, ymm6
	vaddps	ymm0, ymm0, ymm3
	add	rax, 16
	add	r11, 2
	jne	LBB2_12
## %bb.4:
	test	r9, r9
	jne	LBB2_5
	jmp	LBB2_6
LBB2_1:
	vxorps	xmm1, xmm1, xmm1
	vxorps	xmm2, xmm2, xmm2
	vxorps	xmm0, xmm0, xmm0
	jmp	LBB2_6
LBB2_3:
	vxorps	xmm0, xmm0, xmm0
	xor	eax, eax
	vxorps	xmm2, xmm2, xmm2
	vxorps	xmm1, xmm1, xmm1
	test	r9, r9
	je	LBB2_6
LBB2_5:
	vmovups	ymm3, ymmword ptr [rsi + 4*rax]
	vmovups	ymm4, ymmword ptr [rdx + 4*rax]
	vmulps	ymm5, ymm4, ymm4
	vaddps	ymm0, ymm0, ymm5
	vmulps	ymm5, ymm3, ymm3
	vaddps	ymm2, ymm2, ymm5
	vmulps	ymm3, ymm3, ymm4
	vaddps	ymm1, ymm1, ymm3
LBB2_6:
	vhaddps	ymm1, ymm1, ymm1
	vhaddps	ymm1, ymm1, ymm1
	vextractf128	xmm3, ymm1, 1
	vaddss	xmm1, xmm1, xmm3
	vhaddps	ymm2, ymm2, ymm2
	vmovss	dword ptr [rcx], xmm1
	vhaddps	ymm2, ymm2, ymm2
	vextractf128	xmm3, ymm2, 1
	vhaddps	ymm4, ymm0, ymm0
	vaddss	xmm0, xmm2, xmm3
	vhaddps	ymm2, ymm4, ymm4
	vextractf128	xmm3, ymm2, 1
	vaddss	xmm2, xmm2, xmm3
	movsxd	rax, r10d
	cmp	rax, rdi
	jae	LBB2_10
## %bb.7:
	movsxd	rax, r10d
	movsxd	r9, edi
	not	r9
	or	r9, 7
	test	dil, 1
	je	LBB2_9
## %bb.8:
	vmovss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, dword ptr [rdx + 4*rax]
	vaddss	xmm1, xmm1, xmm3
	vmovss	dword ptr [rcx], xmm1
	vmovss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm0, xmm0, xmm3
	vmovss	xmm3, dword ptr [rdx + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm2, xmm2, xmm3
	or	rax, 1
LBB2_9:
	add	r9, rdi
	je	LBB2_10
	.p2align	4, 0x90
LBB2_13:                                ## =>This Inner Loop Header: Depth=1
	vmovss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, dword ptr [rdx + 4*rax]
	vaddss	xmm1, xmm1, xmm3
	vmovss	dword ptr [rcx], xmm1
	vmovss	xmm3, dword ptr [rsi + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmovss	xmm4, dword ptr [rsi + 4*rax + 4] ## xmm4 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm0, xmm0, xmm3
	vmovss	xmm3, dword ptr [rdx + 4*rax] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm2, xmm2, xmm3
	vmulss	xmm3, xmm4, dword ptr [rdx + 4*rax + 4]
	vaddss	xmm1, xmm1, xmm3
	vmovss	dword ptr [rcx], xmm1
	vmovss	xmm3, dword ptr [rsi + 4*rax + 4] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm0, xmm0, xmm3
	vmovss	xmm3, dword ptr [rdx + 4*rax + 4] ## xmm3 = mem[0],zero,zero,zero
	vmulss	xmm3, xmm3, xmm3
	vaddss	xmm2, xmm2, xmm3
	add	rax, 2
	cmp	rdi, rax
	jne	LBB2_13
LBB2_10:
	vmulss	xmm0, xmm0, xmm2
	vmovss	dword ptr [r8], xmm0
	mov	rsp, rbp
	pop	rbp
	vzeroupper
	ret
                                        ## -- End function

.subsections_via_symbols