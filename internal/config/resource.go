package config 


type Executor interface {
	Check()
	Plan()
	Apply()
	Verify()
}


func Order(resources []Resource) []Resource {
	indegree := make([]int, len(resources))
	dependents := make([][]int, len(resources))
	index := make(map[string]int)


	for i, resource := range resources {
		index[resource.ID] = i
	}

	for i, resource := range resources {
		for _, dependency := range resource.Dependencies {
			indegree[i] += 1
			j := index[dependency]
			dependents[j] = append(dependents[j], i)
		}
	}


	queue := make([]int, 0, len(resources))


	for i, val := range indegree {
		if val == 0{
			queue = append(queue, i)
		}
	}


	visit := make(map[int]struct{})

	order := make([]Resource, 0, len(resources))

	for len (queue) > 0{
		i := queue[0]
		queue = queue[1:]

		order  = append(order, resources[i])

		for _, reliant := range dependents[i] {
			indegree[reliant] -= 1

			if indegree[reliant] == 0 {
				queue = append(queue, reliant)
			}
		}

		visit[i] = struct{}{}
	}


	return order
}


