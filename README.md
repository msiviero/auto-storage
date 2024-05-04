A small utility to write persistible structs (with pebble) from proto3 definitions

Compact on db creation
/*
func Compact(d *pebble.DB){
	iter,err := d.NewIter(&pebble.IterOptions{})
	if err!=nil{
		return
	}
var first, last []byte
if iter.First() {
  first = append(first, iter.Key()...)
}
if iter.Last() {
  last = append(last, iter.Key()...)
}
if err := iter.Close(); err != nil {
  return 
}
if err := d.Compact(first, last, true); err != nil {
  return 
}
}
*/

- generate lib.go and lib.proto so it can use helper functions
- handle relations
- handle search by field
