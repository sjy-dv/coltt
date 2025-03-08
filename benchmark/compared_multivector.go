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
	"playground.benchmark.coltt/experimentalproto"
)

func main() {
	// reviews := []map[string]string{
	// 	{
	// 		"recommend_group": "This product is highly recommended by MD for individuals with dry and sensitive skin struggling with hydration.",
	// 		"product_caution": "It features a lightweight gel texture that absorbs quickly for a moisturized finish; however, caution is advised when applied to delicate areas like the eyes.",
	// 		"user_review":     "After use, my skin felt instantly hydrated with a lasting moisturizing effect. It is gentle enough for daily use and safe for sensitive skin.",
	// 	},
	// 	{
	// 		"recommend_group": "MD endorses this product for those experiencing dryness and irritation, providing essential hydration.",
	// 		"product_caution": "The formula is light and absorbs fast, but users should avoid applying it too close to the eye area to prevent irritation.",
	// 		"user_review":     "I noticed a quick boost in hydration after using this product, and my skin stayed moisturized throughout the day.",
	// 	},
	// 	{
	// 		"recommend_group": "Recommended by MD for users with delicate skin needing extra moisture and soothing care.",
	// 		"product_caution": "With a fast-absorbing gel consistency, it provides a non-greasy finish; however, it should be used cautiously on extremely sensitive spots.",
	// 		"user_review":     "It delivered immediate hydration and kept my skin soft and supple for hours, making it a great daily moisturizer.",
	// 	},
	// 	{
	// 		"recommend_group": "This product comes highly recommended by MD for its ability to hydrate dry and sensitive skin.",
	// 		"product_caution": "Its lightweight, gel-based formula ensures quick absorption, but avoid contact with eyes.",
	// 		"user_review":     "Using this product made my skin feel refreshed and continuously moisturized throughout the day.",
	// 	},
	// 	{
	// 		"recommend_group": "MD recommends this product for individuals with dry skin issues, ensuring a boost of hydration when needed.",
	// 		"product_caution": "The gel texture absorbs rapidly, though it's best used with caution on vulnerable skin areas.",
	// 		"user_review":     "I experienced a significant hydration boost immediately after application, with long-lasting moisture that kept my skin comfortable.",
	// 	},
	// 	{
	// 		"recommend_group": "For those battling dry and sensitive skin, MD highlights this product as a top hydration solution.",
	// 		"product_caution": "The product has a refreshing gel texture that is quickly absorbed; users should be mindful of applying near the eyes.",
	// 		"user_review":     "My skin felt remarkably hydrated right away, and the moisturizing effect persisted well into the day.",
	// 	},
	// 	{
	// 		"recommend_group": "MD favors this product for individuals with skin that requires immediate and enduring hydration.",
	// 		"product_caution": "It boasts a light, gel-like texture that offers rapid absorption, yet care should be taken on delicate facial areas.",
	// 		"user_review":     "It instantly quenched my skin's thirst and maintained a soothing moisture level, making it a reliable choice.",
	// 	},
	// 	{
	// 		"recommend_group": "This product is a recommended choice by MD for those in need of a quick hydration fix for sensitive skin.",
	// 		"product_caution": "Its gel formula is both light and fast-absorbing, though extra care is advised for areas around the eyes.",
	// 		"user_review":     "After using it, my skin was visibly hydrated, and the moisture seemed to last throughout the day without any heaviness.",
	// 	},
	// 	{
	// 		"recommend_group": "MD's recommendation for anyone with dryness and sensitivity, this product focuses on delivering essential moisture.",
	// 		"product_caution": "It features a fast-absorbing gel consistency, ensuring a non-oily finish; however, it should be applied carefully near sensitive zones.",
	// 		"user_review":     "The product provided a swift hydration boost and maintained a steady moisture level, ideal for daily routines.",
	// 	},
	// 	{
	// 		"recommend_group": "MD highly recommends this product for dry and sensitive skin, emphasizing its hydrating benefits.",
	// 		"product_caution": "It uses a gel texture for quick absorption and a light finish, but it's important to avoid overuse on very delicate skin areas.",
	// 		"user_review":     "I found that it gave my skin an immediate moisture lift and kept it hydrated for a long period, making it a staple in my routine.",
	// 	},
	// }

	// findKorean := map[string]string{
	// 	"recommend_group": "저는 건조하고 민감한 피부를 가진 사람입니다. 제 피부에 즉각적인 수분 공급과 안정적인 보습 효과를 주는 제품을 원해요.",
	// 	"product_caution": "젤 타입처럼 가볍게 발리면서 빠르게 흡수되는 제품이면 좋겠어요. 특히 눈가 등 민감한 부위에도 자극이 없었으면 합니다.",
	// 	"user_review":     "사용 후 피부가 촉촉해지고 오랜 시간 보습이 유지되는 제품",
	// }

	findQuery := map[string]string{
		"recommend_group": "I have dry skin—any decent product suggestions out there? Not sure if I should trust MD recommendations fully.",
		"product_caution": "Not sure if I prefer a gel or a cream, but I need something lightweight and non-irritating.",
		"user_review":     "I'm curious what actual user reviews say, though I worry they might be a bit overhyped.",
	}

	InitEmbeddings()
	collectionName := "comparedmultivectorscore8"
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dclient := experimentalproto.NewExperimentalMultiVectorRpcClient(conn)
	i, ex := dclient.LoadCollection(context.Background(), &experimentalproto.CollectionName{
		CollectionName: collectionName,
	})
	fmt.Println(i, i.CollectionSize, ex, "1111111111111111")
	// resp2, err := dclient.CreateCollection(context.Background(), &experimentalproto.Collection{
	// 	CollectionName: collectionName,
	// 	Index: []*experimentalproto.Index{
	// 		{
	// 			IndexName:  "group_vector",
	// 			IndexType:  experimentalproto.IndexType_Vector,
	// 			EnableNull: false,
	// 		},
	// 		{
	// 			IndexName:  "caution_vector",
	// 			IndexType:  experimentalproto.IndexType_Vector,
	// 			EnableNull: false,
	// 		},
	// 		{
	// 			IndexName:  "review_vector",
	// 			IndexType:  experimentalproto.IndexType_Vector,
	// 			EnableNull: false,
	// 		},
	// 		{
	// 			IndexName:  "recommend_group",
	// 			IndexType:  experimentalproto.IndexType_String,
	// 			EnableNull: false,
	// 		},
	// 		{
	// 			IndexName:  "product_caution",
	// 			IndexType:  experimentalproto.IndexType_String,
	// 			EnableNull: false,
	// 		},
	// 		{
	// 			IndexName:  "user_review",
	// 			IndexType:  experimentalproto.IndexType_String,
	// 			EnableNull: false,
	// 		},
	// 	},
	// 	Distance:     experimentalproto.Distance_Cosine,
	// 	Quantization: experimentalproto.Quantization_None,
	// 	Dim:          384,
	// 	Versioning:   true,
	// })
	// fmt.Println(resp2)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for i, review := range reviews {
	// 	// mergeText := fmt.Sprintf("recommend_group: %s\nproduct_caution: %s\nuser_review: %s",
	// 	// 	review["recommend_group"], review["product_caution"], review["user_review"])
	// 	vec1, _ := TextEmbedding(review["recommend_group"])
	// 	vec2, _ := TextEmbedding(review["product_caution"])
	// 	vec3, _ := TextEmbedding(review["user_review"])
	// 	m, _ := structpb.NewStruct(map[string]interface{}{
	// 		// "group_vector":    vec1,
	// 		// "caution_vector":  vec2,
	// 		// "review_vector":   vec3,
	// 		"recommend_group": review["recommend_group"],
	// 		"product_caution": review["product_caution"],
	// 		"user_review":     review["user_review"],
	// 	})
	// 	fmt.Println(m)
	// 	resp3, err := dclient.Index(context.Background(), &experimentalproto.IndexChange{
	// 		Id:             strconv.Itoa(i),
	// 		CollectionName: collectionName,
	// 		Changed:        experimentalproto.IndexChagedType_CHANGED,
	// 		Metadata:       m,
	// 		Vectors: []*experimentalproto.VectorIndex{
	// 			{
	// 				IndexName: "group_vector",
	// 				Vector:    vec1,
	// 			},
	// 			{
	// 				IndexName: "caution_vector",
	// 				Vector:    vec2,
	// 			},
	// 			{
	// 				IndexName: "review_vector",
	// 				Vector:    vec3,
	// 			},
	// 		},
	// 	})
	// 	fmt.Println(resp3)
	// 	if err != nil {
	// 		log.Fatal(err, "222222222222")
	// 	}
	// }

	vec1, _ := TextEmbedding(findQuery["recommend_group"])
	vec2, _ := TextEmbedding(findQuery["product_caution"])
	vec3, _ := TextEmbedding(findQuery["user_review"])

	resp, err := dclient.VectorSearch(context.Background(), &experimentalproto.SearchMultiIndex{
		CollectionName:        collectionName,
		TopK:                  10,
		HighResourceAvaliable: true,
		Vector: []*experimentalproto.MultiVectorIndex{
			{
				IndexName:    "group_vector",
				Vector:       vec1,
				IncludeOrNot: true,
				Ratio:        30,
			},
			{
				IndexName:    "caution_vector",
				Vector:       vec2,
				IncludeOrNot: true,
				Ratio:        40,
			},
			{
				IndexName:    "review_vector1",
				Vector:       vec3,
				IncludeOrNot: true,
				Ratio:        30,
			},
		},
	})
	fmt.Println(resp, err, "33333333333333333")
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

// 75.62031 fields:{key:"product_caution"  value:{string_value:"It features a lightweight gel texture that absorbs quickly for a moisturized finish; however, caution is advised when applied to delicate areas like the eyes."}}  fields:{key:"recommend_group"  value:{string_value:"This product is highly recommended by MD for individuals with dry and sensitive skin struggling with hydration."}}  fields:{key:"user_review"  value:{string_value:"After use, my skin felt instantly hydrated with a lasting moisturizing effect. It is gentle enough for daily use and safe for sensitive skin."}}
// 74.87181 fields:{key:"product_caution"  value:{string_value:"The gel texture absorbs rapidly, though it's best used with caution on vulnerable skin areas."}}  fields:{key:"recommend_group"  value:{string_value:"MD recommends this product for individuals with dry skin issues, ensuring a boost of hydration when needed."}}  fields:{key:"user_review"  value:{string_value:"I experienced a significant hydration boost immediately after application, with long-lasting moisture that kept my skin comfortable."}}
// 74.46818 fields:{key:"product_caution"  value:{string_value:"It uses a gel texture for quick absorption and a light finish, but it's important to avoid overuse on very delicate skin areas."}}  fields:{key:"recommend_group"  value:{string_value:"MD highly recommends this product for dry and sensitive skin, emphasizing its hydrating benefits."}}  fields:{key:"user_review"  value:{string_value:"I found that it gave my skin an immediate moisture lift and kept it hydrated for a long period, making it a staple in my routine."}}
// 73.2556 fields:{key:"product_caution"  value:{string_value:"Its lightweight, gel-based formula ensures quick absorption, but avoid contact with eyes."}}  fields:{key:"recommend_group"  value:{string_value:"This product comes highly recommended by MD for its ability to hydrate dry and sensitive skin."}}  fields:{key:"user_review"  value:{string_value:"Using this product made my skin feel refreshed and continuously moisturized throughout the day."}}
// 72.46799 fields:{key:"product_caution"  value:{string_value:"It boasts a light, gel-like texture that offers rapid absorption, yet care should be taken on delicate facial areas."}}  fields:{key:"recommend_group"  value:{string_value:"MD favors this product for individuals with skin that requires immediate and enduring hydration."}}  fields:{key:"user_review"  value:{string_value:"It instantly quenched my skin's thirst and maintained a soothing moisture level, making it a reliable choice."}}
// 72.41062 fields:{key:"product_caution"  value:{string_value:"With a fast-absorbing gel consistency, it provides a non-greasy finish; however, it should be used cautiously on extremely sensitive spots."}}  fields:{key:"recommend_group"  value:{string_value:"Recommended by MD for users with delicate skin needing extra moisture and soothing care."}}  fields:{key:"user_review"  value:{string_value:"It delivered immediate hydration and kept my skin soft and supple for hours, making it a great daily moisturizer."}}
// 70.91707 fields:{key:"product_caution"  value:{string_value:"The product has a refreshing gel texture that is quickly absorbed; users should be mindful of applying near the eyes."}}  fields:{key:"recommend_group"  value:{string_value:"For those battling dry and sensitive skin, MD highlights this product as a top hydration solution."}}  fields:{key:"user_review"  value:{string_value:"My skin felt remarkably hydrated right away, and the moisturizing effect persisted well into the day."}}
// 70.79787 fields:{key:"product_caution"  value:{string_value:"It features a fast-absorbing gel consistency, ensuring a non-oily finish; however, it should be applied carefully near sensitive zones."}}  fields:{key:"recommend_group"  value:{string_value:"MD's recommendation for anyone with dryness and sensitivity, this product focuses on delivering essential moisture."}}  fields:{key:"user_review"  value:{string_value:"The product provided a swift hydration boost and maintained a steady moisture level, ideal for daily routines."}}
// 70.213646 fields:{key:"product_caution"  value:{string_value:"Its gel formula is both light and fast-absorbing, though extra care is advised for areas around the eyes."}}  fields:{key:"recommend_group"  value:{string_value:"This product is a recommended choice by MD for those in need of a quick hydration fix for sensitive skin."}}  fields:{key:"user_review"  value:{string_value:"After using it, my skin was visibly hydrated, and the moisture seemed to last throughout the day without any heaviness."}}
// 68.00476 fields:{key:"product_caution"  value:{string_value:"The formula is light and absorbs fast, but users should avoid applying it too close to the eye area to prevent irritation."}}  fields:{key:"recommend_group"  value:{string_value:"MD endorses this product for those experiencing dryness and irritation, providing essential hydration."}}  fields:{key:"user_review"  value:{string_value:"I noticed a quick boost in hydration after using this product, and my skin stayed moisturized throughout the day."}}
