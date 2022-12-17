<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <meta http-equiv="X-UA-Compatible" content="IE=edge" />
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
        <meta name="description" content="" />
        <meta name="author" content="" />
        <title>Model for {{.Name}}</title>
        <link href="https://ajax.googleapis.com/ajax/libs/jqueryui/1.11.2/themes/smoothness/jquery-ui.css" rel="stylesheet" />
        <link rel="apple-touch-icon" sizes="180x180" href="assets/ico/apple-touch-icon.png">
        <link rel="icon" type="image/png" sizes="32x32" href="assets/ico/favicon-32x32.png">
        <link rel="icon" type="image/png" sizes="16x16" href="assets/ico/favicon-16x16.png">
        <link rel="manifest" href="site.webmanifest">
        <link rel="mask-icon" href="assets/ico/safari-pinned-tab.svg" color="#5bbad5">
        <meta name="msapplication-TileColor" content="#da532c">
        <meta name="theme-color" content="#ffffff">
    </head>
    <style>
      /* Icons from iconfinder.com
http://www.iconfinder.com/search/?q=iconset%3Asiena
*/

* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

html, body {
	height: 100%;
	width: 100%;
}

body {
	font-family: Arial, sans-serif;
	font-size: 14px;
	color: #444;
	background-color: #DDD;
}

::selection {
	background: transparent;
}
::-moz-selection {
	background: transparent;
}

#view {
	width: 100%;
	height: 100%;
	position: relative;
	z-index: 3;
}

.ent {
	display: block;
	float: left;
	position: relative;
	font-size: 12px;
	margin: 20px;
	padding: 5px 10px 5px 5px;
	min-width: 100px;
	background-color: #fff;
	-moz-border-radius: 3px;
	-webkit-border-radius: 3px;
	border-radius: 3px;
	-webkit-box-shadow: 0 1px 3px #666;
	-moz-box-shadow: 0 1px 3px #666;
	box-shadow: 0 1px 3px #666;
	z-index: 2;
	cursor: move;
}

.ent:before {
	content: attr(id);
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	padding: 5px 5px 5px 10px;
	font-size: 12px;
	font-weight: bold;
	color: #eee;
	border-top: 1px solid #CCC;
	background-color: #444;
	background: -webkit-linear-gradient(#444, #333);
	background: -moz-linear-gradient(#444, #333);
	background: -o-linear-gradient(#444, #333);
	background: -ms-linear-gradient(#444, #333);
	background: linear-gradient(#444, #333);
	-moz-border-radius-top: 3px 3px 0px 0px;
	-webkit-border-radius: 3px 3px 0px 0px;
	border-radius: 3px 3px 0px 0px;
}

.ent ul {
	list-style: none;
	margin: 5px 10px 5px 5px;
}

.ent ul li {
	color: #666;
	padding-top: 1px;
	line-height: 16px;
	background-image: url(https://dl.dropbox.com/s/qo412wmz99jjfjw/siena.png?dl=1);
	background-position: 0 0;
	background-color: transparent;
	background-repeat: no-repeat;
	padding-left: 20px;
}

.ent ul.pk {
	margin-top: 30px;
}

.ent ul.pk li {
	background-position: 0 -16px;
	color: #C24704;
}

.ent ul.fk li {
	background-position: 0 -32px;
	color: #8FA63C;
}

.rel {
	position: absolute;
}

.relLT {
	border-left: 1px solid #000;
	border-top: 1px solid #000;
}

.relLB {
	border-left: 1px solid #000;
	border-bottom: 1px solid #000;
}

.relRT {
	border-right: 1px solid #000;
	border-top: 1px solid #000;
}

.relRB {
	border-right: 1px solid #000;
	border-bottom: 1px solid #000;
}

    </style>
    <body>
      <p>Using source code for ER diagram generation from John @Nelms CodePen</p>
      
      {{.Content}}
      <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.1.3/jquery.min.js" crossorigin="anonymous"></script>
      <script src="https://ajax.googleapis.com/ajax/libs/jqueryui/1.11.2/jquery-ui.min.js" crossorigin="anonymous"></script>
      <script language="javascript">
			$(".ent").draggable({   
            	containment: 'body',   
            	scroll: false 
     		});
         		
			var lmb = false;
			
			$(".ent").mousedown(function(e){
    			if(e.which === 1) lmb = true;
			});
			
			$(".ent").mouseup(function(e){
    			if(e.which === 1) lmb = false;
			});
         		
			$(".ent").mousemove(function(e) {
	        	if (e.which === 1 && lmb) {
	        		calculate();
	        	}
    		});
         		
         		calculate = function() {
         		
	         		$(".ent").each(function() {
	         			var ent = $(this);
	         			if ($(this).children(".fk").length != 0) {
	         				$(this).children(".fk").children("li").each(function() {
	         					var tbName = $(this).attr("fk");
	         					var rt = $("#"+tbName);
	         					
	         					if ($("#rel"+tbName).length == 0) {
	         						var rel = $("<div class='rel'></div>");
									rel.attr("id", "rel"+tbName)
	         					}
	         					else
	         						var rel = $("#rel"+tbName);
	         					
	         					var sx = ent.offset().left + Math.round(ent.outerWidth() /2);
	         					var sy = ent.offset().top  + Math.round(ent.outerHeight()/2);
	         					var ex = rt .offset().left + Math.round(rt .outerWidth() /2);
	         					var ey = rt .offset().top  + Math.round(rt .outerHeight()/2);
	         					
	         					var t,l;
	         					if (sy > ey) t = ey; else t = sy;
	         					if (sx > ex) {
	         						l = ex;
	         					} else {
									l = sx;
								}
								
								var cx = sx - ex;
								var cy = sy - ey;
								if ((cx>=0) && (cy>=0)) {
									rel.toggleClass("relRT", true);
									rel.toggleClass("relLB relRB relLT", false);
								} else if ((cx>0) && (cy<0)) {
									rel.toggleClass("relRB", true);
									rel.toggleClass("relLB relLT relRT", false);									
								} else if ((cx<0) && (cy>0)) {
									rel.toggleClass("relLT", true);
									rel.toggleClass("relLB relRT relRB", false);
								} else if ((cx<0) && (cy<0)) {
									rel.toggleClass("relLB", true);
									rel.toggleClass("relLT relRT relRB", false);
								}

	         					rel.offset({top: t, left: l});
	         					rel.height(Math.abs(ey - sy));
	         					rel.width (Math.abs(ex - sx));

	         					$("body").append(rel);
	         				});
	         			}
	         		});
         		}
         		
         		calculate();
      </script>
    </body>
  </html>