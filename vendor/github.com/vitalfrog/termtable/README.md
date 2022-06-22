# termtable

A simple golang terminal table drawing library which allows you to populate the table row by row.

This is how your terminal tables actually could look ;)
```bash
   Path                            | Country    | Device        | Max First Input Delay    | Server response time    | Time to interactive    | Cumulative Layout Shift                     | Largest Contentful Paint                   
=========================================================================================================================================================================================================================================
   /                               | us         | mobile        | 4003ms                   | 16ms                    | 0.000000               | 379ms                                       | 1245ms                                     
                                   |            |               |                          |                         |                        |                                             | div.hero-image                             
                                   |            |               |                          |                         |                        |                                             |  >div.tv-container                         
                                   |            |               |                          |                         |                        |                                             |   >picture                                 
                                   |            |               |                          |                         |                        |                                             |    >img                                    
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
   /de                             | us         | mobile        | 2184ms                   | 16ms                    | 0.000000               | 563ms                                       | 1325ms                                     
                                   |            |               |                          |                         |                        |                                             | a.title-list-grid__item--link              
                                   |            |               |                          |                         |                        |                                             |  >div.title-poster                         
                                   |            |               |                          |                         |                        |                                             |   >picture.picture-comp                    
                                   |            |               |                          |                         |                        |                                             |    >img.picture-comp__img                  
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
   /                               | us         | desktop       | 3795ms                   | 16ms                    | 0.000000               | 118ms                                       | 1327ms                                     
                                   |            |               |                          |                         |                        |                                             | div.hero-image                             
                                   |            |               |                          |                         |                        |                                             |  >div.tv-container                         
                                   |            |               |                          |                         |                        |                                             |   >picture                                 
                                   |            |               |                          |                         |                        |                                             |    >img                                    
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
   /de                             | us         | desktop       | 2704ms                   | 16ms                    | 0.000000               | 124ms                                       | 1328ms                                     
                                   |            |               |                          |                         |                        |                                             | a.title-list-grid__item--link              
                                   |            |               |                          |                         |                        |                                             |  >div.title-poster                         
                                   |            |               |                          |                         |                        |                                             |   >picture.picture-comp                    
                                   |            |               |                          |                         |                        |                                             |    >img.picture-comp__img                  
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

```

## Motivation

All the other table writer libraries I found always required me to have all the data at the moment of drawing. As I use
[JSONL](https://github.com/vitalfrog/jsonl) in the API the data comes one record after another, which made the other libraries
not a fit for the [VitalFrog](https://www.vitalfrog.com/) go client.

## Simple example

```bash
  Path       | Country| Device    
==================================
  /          | us     | mobile    
----------------------------------
  /          | us     | desktop   
----------------------------------
  /de        | de     | mobile    
----------------------------------
```

You can print this beautiful table with the below code example. (Errors dropped for brevity. You should check them!)

```go
package main

import (
	"github.com/vitalfrog/termtable"
	"os"
)

func main() {
	// Creates a new termtable instance with " | " as spacing between columns
	tt := termtable.New(os.Stdout," | ")

	tt.WriteHeader([]termtable.HeaderField{
		{
			Field: termtable.NewStringField("Path"),
		},
		{
			Field: termtable.NewStringField("Country"),
			Width: termtable.IntPointer(4),
		},
		{
			Field: termtable.NewStringField("Device"),
			Width: termtable.IntPointer(10),
		},
	})

	tt.WriteRowDivider('=')
	tt.WriteRow([]termtable.Field{
		termtable.NewStringField("/"),
		termtable.NewStringField("us"),
		termtable.NewStringField("mobile"),
	})
	tt.WriteRowDivider('-')
	tt.WriteRow([]termtable.Field{
		termtable.NewStringField("/"),
		termtable.NewStringField("us"),
		termtable.NewStringField("desktop"),
	})
	tt.WriteRowDivider('-')
	tt.WriteRow([]termtable.Field{
		termtable.NewStringField("/de"),
		termtable.NewStringField("de"),
		termtable.NewStringField("mobile"),
	})
	tt.WriteRowDivider('-')
}
```

## License

MIT
