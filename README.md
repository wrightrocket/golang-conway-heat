# golang-conway-heat
Conway's Game of Life written in Go with heat map color to show neighbor density 


<pre>Usage of ./go-gl-conway-heat:
  -c	Same as -color. (default true)
  -color
    	If true, the number of neighbors a live cell is colored:
      red &gt; 3, yellow = 3, green = 2, and blue &lt; 2.
      If false, then live cells will appear white. (default true)
  -d string
    	Same as -delay. (default &quot;5s&quot;)
  -delay string
    	Sets the amount of time to delay at the end of the game. (default &quot;5s&quot;)
  -e string
    	Same as -expire. (default &quot;0d0h0m0s&quot;)
  -expire string
    	Sets the amount of time to run the game. When -expire is a zero duration, 
      it removes any time constraint. (default &quot;0d0h0m0s&quot;)
  -f int
    	Same as -fps. (default 5)
  -fps int
    	Sets the frames-per-second, used set the speed of the simulation. (default 5)
  -g int
    	Same as -grid. (default 100)
  -grid int
    	Sets both the number of rows and columns for the game grid. (default 100)
  -h  
      Same as -help
  -help
      Display these usage options
  -n	Same as -next. (default true)
  -next
    	Boolean to determine if next alive cell is shown as a purple color.  (default true)
  -p float
    	Same as -probability. (default 0.15)
  -probability float
    	A percentage between 0 and 1 used in conjunction to determine if a cell starts alive. 
      For example, 0.15 means each cell has a 15% probability of starting alive. (default 0.15)
  -r int
    	Same as -report.
  -report int
    	Sets the output report. 1: detailed, 2: comma separated, 3: space separated, 
      4: round number and alive percentage. The default is no output.
  -s int
    	Same as -seed. (default <based upon current time>)
  -seed int
    	Sets the starting seed of the game, used to randomize the initial state. 
  -t int
    	Same as -turns
  -turns int
    	Integer for how many turns to execute. When -turns is zero, 
      it removes any constraint on the number of turns.
</pre>
