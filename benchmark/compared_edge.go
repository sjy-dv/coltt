// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
	"playground.benchmark.coltt/edgeproto"
)

func main() {
	reviews := []map[string]string{
		{
			"recommend_group": "This product is highly recommended by MD for individuals with dry and sensitive skin struggling with hydration.",
			"product_caution": "It features a lightweight gel texture that absorbs quickly for a moisturized finish; however, caution is advised when applied to delicate areas like the eyes.",
			"user_review":     "After use, my skin felt instantly hydrated with a lasting moisturizing effect. It is gentle enough for daily use and safe for sensitive skin.",
		},
		{
			"recommend_group": "MD endorses this product for those experiencing dryness and irritation, providing essential hydration.",
			"product_caution": "The formula is light and absorbs fast, but users should avoid applying it too close to the eye area to prevent irritation.",
			"user_review":     "I noticed a quick boost in hydration after using this product, and my skin stayed moisturized throughout the day.",
		},
		{
			"recommend_group": "Recommended by MD for users with delicate skin needing extra moisture and soothing care.",
			"product_caution": "With a fast-absorbing gel consistency, it provides a non-greasy finish; however, it should be used cautiously on extremely sensitive spots.",
			"user_review":     "It delivered immediate hydration and kept my skin soft and supple for hours, making it a great daily moisturizer.",
		},
		{
			"recommend_group": "This product comes highly recommended by MD for its ability to hydrate dry and sensitive skin.",
			"product_caution": "Its lightweight, gel-based formula ensures quick absorption, but avoid contact with eyes.",
			"user_review":     "Using this product made my skin feel refreshed and continuously moisturized throughout the day.",
		},
		{
			"recommend_group": "MD recommends this product for individuals with dry skin issues, ensuring a boost of hydration when needed.",
			"product_caution": "The gel texture absorbs rapidly, though it's best used with caution on vulnerable skin areas.",
			"user_review":     "I experienced a significant hydration boost immediately after application, with long-lasting moisture that kept my skin comfortable.",
		},
		{
			"recommend_group": "For those battling dry and sensitive skin, MD highlights this product as a top hydration solution.",
			"product_caution": "The product has a refreshing gel texture that is quickly absorbed; users should be mindful of applying near the eyes.",
			"user_review":     "My skin felt remarkably hydrated right away, and the moisturizing effect persisted well into the day.",
		},
		{
			"recommend_group": "MD favors this product for individuals with skin that requires immediate and enduring hydration.",
			"product_caution": "It boasts a light, gel-like texture that offers rapid absorption, yet care should be taken on delicate facial areas.",
			"user_review":     "It instantly quenched my skin's thirst and maintained a soothing moisture level, making it a reliable choice.",
		},
		{
			"recommend_group": "This product is a recommended choice by MD for those in need of a quick hydration fix for sensitive skin.",
			"product_caution": "Its gel formula is both light and fast-absorbing, though extra care is advised for areas around the eyes.",
			"user_review":     "After using it, my skin was visibly hydrated, and the moisture seemed to last throughout the day without any heaviness.",
		},
		{
			"recommend_group": "MD's recommendation for anyone with dryness and sensitivity, this product focuses on delivering essential moisture.",
			"product_caution": "It features a fast-absorbing gel consistency, ensuring a non-oily finish; however, it should be applied carefully near sensitive zones.",
			"user_review":     "The product provided a swift hydration boost and maintained a steady moisture level, ideal for daily routines.",
		},
		{
			"recommend_group": "MD highly recommends this product for dry and sensitive skin, emphasizing its hydrating benefits.",
			"product_caution": "It uses a gel texture for quick absorption and a light finish, but it's important to avoid overuse on very delicate skin areas.",
			"user_review":     "I found that it gave my skin an immediate moisture lift and kept it hydrated for a long period, making it a staple in my routine.",
		},
	}

	// findKorean := map[string]string{
	// 	"recommend_group":      "저는 건조하고 민감한 피부를 가진 사람입니다. 제 피부에 즉각적인 수분 공급과 안정적인 보습 효과를 주는 제품을 원해요.",
	// 	"product_caution": "젤 타입처럼 가볍게 발리면서 빠르게 흡수되는 제품이면 좋겠어요. 특히 눈가 등 민감한 부위에도 자극이 없었으면 합니다.",
	// 	"user_review":         "사용 후 피부가 촉촉해지고 오랜 시간 보습이 유지되는 제품",
	// }

	findQuery := map[string]string{
		"recommend_group": "I have dry skin—any decent product suggestions out there? Not sure if I should trust MD recommendations fully.",
		"product_caution": "Not sure if I prefer a gel or a cream, but I need something lightweight and non-irritating.",
		"user_review":     "I'm curious what actual user reviews say, though I worry they might be a bit overhyped.",
	}

	InitEmbeddings()
	collectionName := "compared_edge_score"
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dclient := edgeproto.NewEdgeRpcClient(conn)

	_, err = dclient.CreateCollection(context.Background(), &edgeproto.Collection{
		CollectionName: collectionName,
		Dim:            384,
		Distance:       edgeproto.Distance_Cosine,
		Quantization:   edgeproto.Quantization_None,
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, review := range reviews {
		mergeText := fmt.Sprintf("recommend_group: %s\nproduct_caution: %s\nuser_review: %s",
			review["recommend_group"], review["product_caution"], review["user_review"])
		vec, err := TextEmbedding(mergeText)
		if err != nil {
			log.Fatal(err)
		}
		m, _ := structpb.NewStruct(map[string]interface{}{
			"text": mergeText,
		})

		_, err = dclient.Insert(context.Background(), &edgeproto.ModifyDataset{
			Id:             "test",
			Metadata:       m,
			Vector:         vec,
			CollectionName: collectionName,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	mergeQuery := fmt.Sprintf("recommend_group: %s\nproduct_caution: %s\nuser_review: %s",
		findQuery["recommend_group"], findQuery["product_caution"], findQuery["user_review"])
	vec2, err := TextEmbedding(mergeQuery)
	if err != nil {
		log.Fatal(err)
	}

	resp, _ := dclient.VectorSearch(context.Background(), &edgeproto.SearchReq{
		CollectionName:        collectionName,
		Vector:                vec2,
		TopK:                  10,
		HighResourceAvaliable: true,
	})
	for _, r := range resp.Candidates {
		fmt.Println(r.Score, r.Metadata)
	}
}

const defaultURL = "ws://localhost:8765"

var ec *websocket.Conn

func InitEmbeddings() error {
	c, _, err := websocket.DefaultDialer.Dial(defaultURL, nil)
	if err != nil {
		return err
	}
	ec = c
	return nil
}

func TextEmbedding(text string) ([]float32, error) {
	msgBytes, err := json.Marshal(map[string]string{"sentence": text})
	if err != nil {
		return nil, err
	}
	err = ec.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		return nil, err
	}

	_, msg, err := ec.ReadMessage()
	if err != nil {
		return nil, err
	}
	var embeddings []float32
	err = json.Unmarshal(msg, &embeddings)
	if err != nil {
		return nil, err
	}
	return embeddings, nil
}

// status:true  candidates:{metadata:{}  score:98.27182} <nil>

// 91.79629 fields:{key:"text"  value:{string_value:"recommend_group: MD recommends this product for individuals with dry skin issues, ensuring a boost of hydration when needed.\nproduct_caution: The gel texture absorbs rapidly, though it's best used with caution on vulnerable skin areas.\nuser_review: I experienced a significant hydration boost immediately after application, with long-lasting moisture that kept my skin comfortable."}}
// 91.76757 fields:{key:"text"  value:{string_value:"recommend_group: Recommended by MD for users with delicate skin needing extra moisture and soothing care.\nproduct_caution: With a fast-absorbing gel consistency, it provides a non-greasy finish; however, it should be used cautiously on extremely sensitive spots.\nuser_review: It delivered immediate hydration and kept my skin soft and supple for hours, making it a great daily moisturizer."}}
// 91.67139 fields:{key:"text"  value:{string_value:"recommend_group: MD highly recommends this product for dry and sensitive skin, emphasizing its hydrating benefits.\nproduct_caution: It uses a gel texture for quick absorption and a light finish, but it's important to avoid overuse on very delicate skin areas.\nuser_review: I found that it gave my skin an immediate moisture lift and kept it hydrated for a long period, making it a staple in my routine."}}
// 89.458084 fields:{key:"text"  value:{string_value:"recommend_group: This product comes highly recommended by MD for its ability to hydrate dry and sensitive skin.\nproduct_caution: Its lightweight, gel-based formula ensures quick absorption, but avoid contact with eyes.\nuser_review: Using this product made my skin feel refreshed and continuously moisturized throughout the day."}}
// 89.187386 fields:{key:"text"  value:{string_value:"recommend_group: This product is highly recommended by MD for individuals with dry and sensitive skin struggling with hydration.\nproduct_caution: It features a lightweight gel texture that absorbs quickly for a moisturized finish; however, caution is advised when applied to delicate areas like the eyes.\nuser_review: After use, my skin felt instantly hydrated with a lasting moisturizing effect. It is gentle enough for daily use and safe for sensitive skin."}}
// 88.54189 fields:{key:"text"  value:{string_value:"recommend_group: For those battling dry and sensitive skin, MD highlights this product as a top hydration solution.\nproduct_caution: The product has a refreshing gel texture that is quickly absorbed; users should be mindful of applying near the eyes.\nuser_review: My skin felt remarkably hydrated right away, and the moisturizing effect persisted well into the day."}}
// 87.62555 fields:{key:"text"  value:{string_value:"recommend_group: MD favors this product for individuals with skin that requires immediate and enduring hydration.\nproduct_caution: It boasts a light, gel-like texture that offers rapid absorption, yet care should be taken on delicate facial areas.\nuser_review: It instantly quenched my skin's thirst and maintained a soothing moisture level, making it a reliable choice."}}
// 86.63659 fields:{key:"text"  value:{string_value:"recommend_group: MD's recommendation for anyone with dryness and sensitivity, this product focuses on delivering essential moisture.\nproduct_caution: It features a fast-absorbing gel consistency, ensuring a non-oily finish; however, it should be applied carefully near sensitive zones.\nuser_review: The product provided a swift hydration boost and maintained a steady moisture level, ideal for daily routines."}}
// 86.61372 fields:{key:"text"  value:{string_value:"recommend_group: This product is a recommended choice by MD for those in need of a quick hydration fix for sensitive skin.\nproduct_caution: Its gel formula is both light and fast-absorbing, though extra care is advised for areas around the eyes.\nuser_review: After using it, my skin was visibly hydrated, and the moisture seemed to last throughout the day without any heaviness."}}
// 86.34415 fields:{key:"text"  value:{string_value:"recommend_group: MD endorses this product for those experiencing dryness and irritation, providing essential hydration.\nproduct_caution: The formula is light and absorbs fast, but users should avoid applying it too close to the eye area to prevent irritation.\nuser_review: I noticed a quick boost in hydration after using this product, and my skin stayed moisturized throughout the day."}}
