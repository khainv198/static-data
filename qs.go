	s.router.GET("", func(ctx *gin.Context) {
		str := "?sort[0]=title:asc&sort[1]=content:desc&filters[$or][0][title][$eq]=hello&filters[$or][0][$and][0][a][$eq]=a&filters[$or][0][$and][0][$and][0][a][$eq]=a&filters[$or][0][$and][0][$and][1][b][$eq]=b&filters[$or][0][$and][0][$and][2][c][$eq]=c&filters[$or][0][$and][1][b][$eq]=b&filters[$or][0][$and][1][$and][0][a][$eq]=a&filters[$or][0][$and][1][$and][1][b][$eq]=b&filters[$or][0][$and][1][$and][2][c][$eq]=c&filters[$or][0][$and][2][c][$eq]=c&filters[$or][0][$and][2][$and][0][a][$eq]=a&filters[$or][0][$and][2][$and][1][b][$eq]=b&filters[$or][0][$and][2][$and][2][c][$eq]=c&filters[$or][1][desc][$eq]=hello&filters[$or][1][$and][0][a][$eq]=a&filters[$or][1][$and][1][b][$eq]=b&filters[$or][1][$and][2][c][$eq]=c&filters[$and][0][age][$eq]=10&filters[$and][0][$or][0][a][$eq]=a&filters[$and][0][$or][1][b][$eq]=b&filters[$and][0][$or][2][c][$eq]=c&filters[$and][1][age][$eq]=11&filters[$and][1][$or][0][a][$eq]=a&filters[$and][1][$or][1][b][$eq]=b&filters[$and][1][$or][2][c][$eq]=c&filters[gender][$in][0]=Male&filters[gender][$in][1]=Female&filters[status][$nin][0]=inactive&filters[status][$nin][1]=waiting&filters[age][$gt]=10&filters[age][$lt]=20&filters[title][$ne]=hello&populate[author][sort][0]=title:asc&populate[author][sort][1]=content:desc&populate[author][fields][0]=firstName&populate[author][fields][1]=lastName&populate[author][pagination][offset]=10&populate[author][pagination][page]=1&populate[author][pagination][limit]=10&populate[author][filters][$or][0][title][$eq]=hello&populate[author][filters][$or][1][desc][$eq]=hello&populate[author][filters][$and][0][age][$eq]=10&populate[author][filters][$and][1][age][$eq]=11&populate[author][filters][gender][$in][0]=Male&populate[author][filters][gender][$in][1]=Female&populate[author][filters][status][$nin][0]=inactive&populate[author][filters][status][$nin][1]=waiting&populate[author][populate][profile][sort][0]=title:asc&populate[author][populate][profile][sort][1]=content:desc&populate[author][populate][profile][fields][0]=firstName&populate[author][populate][profile][fields][1]=lastName&populate[author][populate][profile][pagination][offset]=10&populate[author][populate][profile][pagination][page]=1&populate[author][populate][profile][pagination][limit]=10&populate[author][populate][profile][filters][$or][0][title][$eq]=hello&populate[author][populate][profile][filters][$or][1][desc][$eq]=hello&populate[author][populate][profile][filters][$and][0][age][$eq]=10&populate[author][populate][profile][filters][$and][1][age][$eq]=11&populate[author][populate][profile][filters][gender][$in][0]=Male&populate[author][populate][profile][filters][gender][$in][1]=Female&populate[author][populate][profile][filters][status][$nin][0]=inactive&populate[author][populate][profile][filters][status][$nin][1]=waiting&fields[0]=title&fields[1]=content&pagination[offset]=10&pagination[page]=1&pagination[limit]=10"

		groups := map[string][]string{}

		paginationRe := regexp.MustCompile(`(\[?)pagination(\]?)\[(offset|limit|page)\]$`)
		fieldsRe := regexp.MustCompile(`(\[?)fields(\]?)\[\d+\]$`)
		sortRe := regexp.MustCompile(`(\[?)sort(\]?)\[\d+\]$`)

		queries, _ := url.ParseQuery(str)
		for key := range queries {
			if paginationRe.MatchString(key) {
				k := paginationRe.ReplaceAllString(key, "")
				k = k + "[pagination]"

				if _, ok := groups[k]; !ok {
					groups[k] = []string{}
				}

				val, err := strconv.ParseInt(queries[key][0], 10, 64)
				if err != nil {
					continue
				}

				groups[k] = append(groups[k], fmt.Sprintf("%s:%d", regexp.MustCompile(`(offset|limit|page)`).FindString(key), val))
				continue
			}

			if fieldsRe.MatchString(key) {
				k := fieldsRe.ReplaceAllString(key, "")
				k = k + "[fields]"
				if _, ok := groups[k]; !ok {
					groups[k] = []string{}
				}

				groups[k] = append(groups[k], queries[key][0])
				continue
			}

			if sortRe.MatchString(key) {
				k := sortRe.ReplaceAllString(key, "")
				if k == "" {
					k = "sort"
				} else {
					k = k + "[sort]"
				}

				arr := strings.Split(strings.ToLower(queries[key][0]), ":")
				if len(arr) != 2 {
					continue
				}

				var val string

				switch arr[1] {
				case "asc", "1":
					val = "asc"
				case "desc", "-1":
					val = "desc"
				default:
					val = ""
				}

				if val == "" {
					continue
				}

				if _, ok := groups[k]; !ok {
					groups[k] = []string{}
				}

				groups[k] = append(groups[k], fmt.Sprintf("%s:%s", arr[0], val))
				continue
			}

			if strings.Contains(key, "filters") {
				filterIdx := strings.Index(key, "filters")
				filterKey := strings.TrimPrefix(strings.TrimPrefix(key[filterIdx:], "filters"), "]")
				prefix := strings.TrimSuffix(key[:filterIdx], "[")
				prefix = prefix + "[filters]"

				if _, ok := groups[prefix]; !ok {
					groups[prefix] = []string{}
				}

				groups[prefix] = append(groups[prefix], fmt.Sprintf("%s:%s", filterKey, queries[key][0]))
			}
		}

		populateRe := regexp.MustCompile(`\[(pagination|sort|filters|fields)\]$`)
		populateGroups := map[string]bool{}
		for key := range groups {
			if strings.Contains(key, "populate") {
				k := populateRe.ReplaceAllString(key, "")
				if populateGroups[k] {
					continue
				}

				populateGroups[k] = true
			}
		}
		populateGroups[""] = true

		for key := range populateGroups {
			filters := groups[fmt.Sprintf("%s[filters]", key)]
			sort := groups[fmt.Sprintf("%s[sort]", key)]
			fields := groups[fmt.Sprintf("%s[fields]", key)]
			pagination := groups[fmt.Sprintf("%s[pagination]", key)]

			filterQuery := bson.M{}
			for _, filter := range filters {

				arr := strings.Split(filter, ":")
				var cur interface{}
				elements := strings.Split(strings.TrimSuffix(strings.TrimPrefix(arr[0], "["), "]"), "][")
				for index, ele := range elements {
					if index == 0 {
						cur = nil
						if ele == "$and" || ele == "$or" {
							idx, _ := strconv.Atoi(elements[index+1])
							if _, ok := filterQuery[ele]; !ok {
								filterQuery[ele] = []bson.M{}
							}

							for i := 0; i <= idx; i++ {
								if idx >= len(filterQuery[ele].([]bson.M)) {
									filterQuery[ele] = append(filterQuery[ele].([]bson.M), bson.M{})
								}
							}

							cur = filterQuery[ele].([]bson.M)[idx]
						} else {
							if _, ok := filterQuery[ele]; !ok {
								filterQuery[ele] = bson.M{}
							}

							cur = filterQuery[ele]
						}
					}

					if _, err := strconv.Atoi(ele); err == nil {
						continue
					}

					if (ele == "$and" || ele == "$or") && index > 0 {
						if _, ok := cur.(bson.M)[ele]; !ok {
							cur.(bson.M)[ele] = []bson.M{}
						}

						idx, _ := strconv.Atoi(elements[index+1])

						for i := 0; i <= idx; i++ {
							if idx >= len(cur.(bson.M)[ele].([]bson.M)) {
								cur.(bson.M)[ele] = append(cur.(bson.M)[ele].([]bson.M), bson.M{})
							}
						}

						cur = cur.(bson.M)[ele].([]bson.M)[idx]
					}

					if (ele == "$in" || ele == "$nin") && index > 0 {
						if cur.(bson.M)[ele] == nil {
							cur.(bson.M)[ele] = []interface{}{}
						}

						idx, _ := strconv.Atoi(elements[index+1])
						for i := 0; i <= idx; i++ {
							if idx >= len(cur.(bson.M)[ele].([]interface{})) {
								var tmp interface{}
								cur.(bson.M)[ele] = append(cur.(bson.M)[ele].([]interface{}), tmp)
							}
						}

						cur.(bson.M)[ele].([]interface{})[idx] = arr[1]
					}

					if !strings.HasPrefix(ele, "$") {
						val := arr[1]
						condition := elements[index+1]
						switch condition {
						case "$gt", "$gte", "$lt", "$lte", "$eq", "$ne":
							if _, ok := cur.(bson.M)[ele]; !ok {
								cur.(bson.M)[ele] = bson.M{}
							}

							cur.(bson.M)[ele].(bson.M)[condition] = val
						}
					}
				}
			}

			sortQuery := bson.M{}
			for _, s := range sort {
				arr := strings.Split(s, ":")
				sortQuery[arr[0]] = arr[1]
			}
			if len(sortQuery) == 0 {
				sortQuery["updatedAt"] = -1
			}

			fieldsQuery := bson.M{}
			for _, f := range fields {
				fieldsQuery[f] = 1
			}
			if len(fieldsQuery) > 0 {
				if _, ok := fieldsQuery["_id"]; !ok {
					fieldsQuery["_id"] = 1
				}
			}

			limit := 10
			page := 1
			offset := 0

			for _, p := range pagination {
				arr := strings.Split(p, ":")
				val, _ := strconv.Atoi(arr[1])
				switch arr[0] {
				case "limit":
					limit = val
				case "page":
					page = val
				case "offset":
					offset = val
				}
			}

			if offset == 0 {
				offset = (page - 1) * limit
			}

			paginationQuery := []bson.M{
				{"$skip": offset},
				{"$limit": limit},
			}

			log.Print(paginationQuery)
		}
	})
