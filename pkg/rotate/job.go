package rotate

type job struct {
  file string
  isAPIKey bool
  oldData []byte
  newData []byte
  bytesWritten int
  name string
  namespace string
  err error
}

type jobs []job

func (j *jobs) add(jbs ...job) {
	for _, curr := range jbs {
		*j = append(*j, curr)
	}
}