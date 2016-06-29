// JavaScript Document
function ResizeHeader(){
	'use strict';
	
	var w = $("body").width();
	var h = document.documentElement.clientHeight;
	$(".navbar-nav").css({"height":h + "px"});
	if (w >= 992 && w < 1500) {
		$("header .logo").prependTo($("header .navbar-nav"));
	} else {
		$("header .logo").prependTo($("header"));
	}
	if (w < 992) {
		$("header .login").insertBefore($("header .navbar-nav ul"));
		$(".mainmenu").css({"top":h - 50 + "px"});
	} else {
		$("header .login").appendTo($(".mainmenu ul"));
		$(".mainmenu").css({"top":"0px"});
	}
}

function HideMenu(){
	'use strict';
	
	if ($("header").hasClass("on")) {
		$("header").removeClass("on");
	}
}

(function($) { 
   $.fn.touchwipe = function(settings) {
     var config = {
    		min_move_x: 50,
 			wipeLeft: function() {},
 			wipeRight: function() {},
			preventDefaultEvents: true
	 };
     
     if (settings) $.extend(config, settings);
 
     this.each(function() {
    	 var startX;
		 var isMoving = false;

    	 function cancelTouch() {
    		 this.removeEventListener('touchmove', onTouchMove);
    		 startX = null;
    		 isMoving = false;
    	 }	
    	 
    	 function onTouchMove(e) {
    		 if(config.preventDefaultEvents) {
    			 //e.preventDefault();
    		 }
    		 if(isMoving) {
	    		 var x = e.touches[0].pageX;
	    		 var dx = startX - x;
	    		 if(Math.abs(dx) >= config.min_move_x) {
	    			cancelTouch();
	    			if(dx > 0) {
	    				config.wipeLeft();
	    			}
	    			else {
	    				config.wipeRight();
	    			}
	    		 }
    		 }
    	 }
    	 
    	 function onTouchStart(e)
    	 {
    		 if (e.touches.length == 1) {
    			 startX = e.touches[0].pageX;
    			 isMoving = true;
    			 this.addEventListener('touchmove', onTouchMove, false);
    		 }
    	 }    	 
    		 
		 this.addEventListener('touchstart', onTouchStart, false);
     });
 
     return this;
   };
 
 })(jQuery);