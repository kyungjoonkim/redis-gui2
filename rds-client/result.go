package rds_client

type Result struct {
	content interface{}
}

func (r *Result) Content() interface{} {
	return r.content
}
