package ui

import "strings"

const ascii_height = 5

// https://onlineasciitools.com/convert-text-to-ascii-art
const ascii_waiting = `                      ##      ##    ###    #### ######## #### ##    ##  ######                       
                      ##  ##  ##  ##   ##   ##     ##     ##  ####  ## ##                            
                      ##  ##  ## ##     ##  ##     ##     ##  ## ## ## ##   ###                      
                      ##  ##  ## #########  ##     ##     ##  ##   ### ##    ##                      
                       ###  ###  ##     ## ####    ##    #### ##    ##  ######                       
`

const ascii_countdown = ` ######   #######  ##     ## ##    ## ######## ########   #######  ##      ## ##    ##       
##    ## ##     ## ##     ## ###   ##    ##    ##     ## ##     ## ##  ##  ## ###   ##  ##   
##       ##     ## ##     ## ##  ####    ##    ##     ## ##     ## ##  ##  ## ##  ####       
##    ## ##     ## ##     ## ##   ###    ##    ##     ## ##     ## ##  ##  ## ##   ###  ##   
 ######   #######   #######  ##    ##    ##    ########   #######   ###  ###  ##    ##       
`

const ascii_colon = `    
 ## 
    
 ## 
    
`

const ascii_one = `   ##   
 ####   
   ##   
   ##   
 ###### 
`
const ascii_two = ` ###### 
     ## 
 ###### 
##      
####### 
`
const ascii_three = `######  
     ## 
 #####  
     ## 
######  
`
const ascii_four = `##      
##  ##  
##  ##  
####### 
    ##  
`
const ascii_five = `###### 
##     
#####  
    ## 
###### 
`
const ascii_six = ` #####  
##      
######  
##   ## 
 #####  
`
const ascii_seven = ` ###### 
##   ## 
    ##  
   ##   
  ##    
`
const ascii_eight = ` #####  
##   ## 
 #####  
##   ## 
 #####  
`
const ascii_nine = ` #####  
##   ## 
 ###### 
     ## 
 #####  
`
const ascii_zero = `  ###   
 ## ##  
##   ## 
 ## ##  
  ###   
`

func concat_ascii(s1 string, s2 string) string {
	ss1 := strings.Split(s1, "\n")
	ss2 := strings.Split(s2, "\n")
	s_out := ""
	for i := 0; i < 5; i++ {
		s_out += ss1[i] + ss2[i] + "\n"
	}
	return s_out
}

func string_to_ascii(s1 string) string {

	number2ascii := map[rune]string{
		'0': ascii_zero,
		'1': ascii_one,
		'2': ascii_two,
		'3': ascii_three,
		'4': ascii_four,
		'5': ascii_five,
		'6': ascii_six,
		'7': ascii_seven,
		'8': ascii_eight,
		'9': ascii_nine,
		':': ascii_colon,
	}

	s_out := `




`
	for _, v := range s1 {
		s_out = concat_ascii(s_out, number2ascii[v])
	}
	return s_out
}

/*
const ascii_one = `
  ##
####
  ##
  ##
  ##
  ##
######
`
const ascii_two = `
 #######
##     ##
       ##
 #######
##
##
#########
`
const ascii_three = `
 #######
##     ##
       ##
 #######
       ##
##     ##
 #######
`
const ascii_four = `
##
##    ##
##    ##
##    ##
#########
      ##
      ##
`
const ascii_five = `
########
##
##
#######
      ##
##    ##
 ######
`
const ascii_six = `
 #######
##     ##
##
########
##     ##
##     ##
 #######
`
const ascii_seven = `
########
##    ##
    ##
   ##
  ##
  ##
  ##
`
const ascii_eight = `
 #######
##     ##
##     ##
 #######
##     ##
##     ##
 #######
`
const ascii_nine = `
 #######
##     ##
##     ##
 ########
       ##
##     ##
 #######
`
const ascii_zero = `
  #####
 ##   ##
##     ##
##     ##
##     ##
 ##   ##
  #####
`
*/
