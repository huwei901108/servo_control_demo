
pakage main

import(

"fmt"

)

const ReadTimeoutInSec=10

func ReadPosWithTime(id byte) (pos int, err error){

	type read_pos struct{
		pos int;
		err error;
	}

	chan_res := make (chan read_pos, 1)
	go func(){
		readpos,readerr := ReadPosition(id)
		chan_res <- read_pos{pos:readpos; err:readerr}
	}()

	select{
		case res:= <-chan_res
	}

}
