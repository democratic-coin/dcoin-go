var coords = {
	archive : {},
	getObject : function(id) {
		if (typeof this.archive[id] == "undefined") {
			this.archive[id] = new this.fc(id);
		}
		return this.archive[id];
	}
};


coords.fc = function(id) { // Constructor

	// Vars
	this.id = document.getElementById(id);
	this.main_area = '';
	this.otstup_x = [];
	this.otstup_y = [];
	this.otstup_x2 = [];
	this.otstup_y2 = [];	
	this.example_coords = [];
	this.coords_array = [];
	this.arr_i = [];
	this.for_mouse_move = '';
	this.user_coords = [];
	this.USER_COORDS = false;
    this.example_area = '';
    this.line_color = '';
    this.type = '';
    this.ugol = [];
}

coords.fc.prototype = {

	getPos: function (id) {

		var elem = document.getElementById(id);
	
		var l = 0;
		var t = 0;
	
		while (elem) {
		
			l += elem.offsetLeft;
			t += elem.offsetTop;
			elem = elem.offsetParent;
		}

		return {"left":l, "top":t};
	},	
	
	cnv_2_coord: function (p1, p2, draw_number, line_color) {
	
		var drawingCanvas = document.getElementById( this.main_area );
				
		if(drawingCanvas && drawingCanvas.getContext) {
		
			var context = drawingCanvas.getContext('2d');
			
			//if (p1[0]<350 && p1[1]<350)
			//{
            if (draw_number  ) {

				context.fillStyle = "#007000";
				context.font = "30pt Arial";
				context.fillText(this.number, p1[0], p1[1]);
			}

			context.beginPath();
			context.lineWidth=1;
			context.moveTo(p1[0], p1[1]);
			context.lineTo(p2[0], p2[1]);
			context.strokeStyle = line_color;
			context.stroke();

		}
	},

	y_line: function (point) {

		this.cnv_2_coord(point, Array(point[0]+this.otstup_y[this.main_area], point[1]-2048), false,  this.line_color);
		this.cnv_2_coord(point, Array(point[0]-this.otstup_y[this.main_area], point[1]+2048), false,  this.line_color);
	},

	x_line: function (point) {

		this.cnv_2_coord(point, Array(point[0]-2048, point[1]-this.otstup_x[this.main_area]), false,  this.line_color);
		this.cnv_2_coord(point, Array(point[0]+2048, point[1]+this.otstup_x[this.main_area]), false,  this.line_color);
	},

	cross: function (point) {

		// x
		this.cnv_2_coord(point, Array(point[0], point[1]-10), true, '#ff0000');
		this.cnv_2_coord(point, Array(point[0], point[1]+10), true, '#ff0000');

		// y
		this.cnv_2_coord(point, Array(point[0]-10, point[1]), true, '#ff0000');
		this.cnv_2_coord(point, Array(point[0]+10, point[1]), true, '#ff0000');

	},

	draw_angle: function (point1, point2, type, ugol) {

		kat1 = point2[0]-point1[0];
		cos1 = Math.cos(ugol*Math.PI/180);
		gep1 = kat1/cos1;
		kat_protiv = Math.sqrt( Math.pow(gep1, 2) - Math.pow(kat1, 2));

		switch (type){

		case 'left_bottom':

			y_glas_nos = point2[1] - point1[1]; // +-
			gep_main1 = y_glas_nos-kat_protiv;  // +-
			katet_protiv_main2 = gep_main1*Math.sin(ugol*Math.PI/180);
			main_GEPO = gep1+katet_protiv_main2; // +-
			main_KATET_PROTIV_Y = main_GEPO*Math.sin(ugol*Math.PI/180);
			main_KATET_PRIL_X = Math.sqrt( Math.pow(main_GEPO, 2) - Math.pow(main_KATET_PROTIV_Y, 2));

			MAIN_X = point2[0]-main_KATET_PRIL_X; // +-
			MAIN_Y = point2[1]-main_KATET_PROTIV_Y; // +-

			BIG_katet = Math.tan((45-ugol)*Math.PI/180)*1500;

			this.cnv_2_coord(Array(MAIN_X, MAIN_Y), Array(MAIN_X+1500, MAIN_Y-BIG_katet), false, this.line_color);

			break;

		case 'right_bottom':

			y_glas_nos = point1[1] - point2[1];
			gep_main1 = y_glas_nos+kat_protiv;
			katet_protiv_main2 = gep_main1*Math.sin(ugol*Math.PI/180);
			main_GEPO = gep1-katet_protiv_main2;
			main_KATET_PROTIV_Y = main_GEPO*Math.sin(ugol*Math.PI/180);
			main_KATET_PRIL_X = Math.sqrt( Math.pow(main_GEPO, 2) - Math.pow(main_KATET_PROTIV_Y, 2));

			MAIN_X = point1[0]+main_KATET_PRIL_X;
			MAIN_Y = point1[1]+main_KATET_PROTIV_Y;

			BIG_katet = Math.tan((45-ugol)*Math.PI/180)*1500;

			this.cnv_2_coord(Array(MAIN_X, MAIN_Y), Array(MAIN_X-1500, MAIN_Y-BIG_katet), false, this.line_color);

			break;

		}

		//this.cross(Array(Math.round(MAIN_X), Math.round(MAIN_Y)));
	},


	init : function(hash) {

		for (var i in hash) this[i] = hash[i];

		if (!this.otstup_x[this.main_area])
			this.otstup_x[this.main_area] = 0;
        if (!this.otstup_y[this.main_area])
            this.otstup_y[this.main_area] = 0;
        if (!this.ugol[this.main_area])
            this.ugol[this.main_area] = 0;


		this.main_area_ = this.main_area;

		// если уже есть массив со всеми точками, то наносим их на рабочую область и заполняем пример
		if ( this.user_coords.length>0 ) {

			this.USER_COORDS = true;

			// нет смещения, т.к. нет курсора мышки
			this.otstup_x2[this.main_area] = 0;
			this.otstup_y2[this.main_area] = 0;

			for (var i=0; i<this.user_coords.length; i++) {

				this.click({pageX:this.user_coords[i][0], pageY:this.user_coords[i][1]});

			}

			this.otstup_x2[this.example_area] = 0;
			this.otstup_y2[this.example_area] = 0;
			for (var i=0; i<this.example_coords.length; i++) {

				this.main_area = this.example_area;
				this.click({pageX:this.example_coords[i][0], pageY:this.example_coords[i][1]});

			}

			this.main_area = this.main_area_;

		}
		else {

			this.first_coord();

			var _this = this;

			var _mousemove = _this.mousemove.bind(_this);
			this.id.addEventListener('mousemove', _mousemove,false);

			var _click = _this.click.bind(_this);
			this.id.addEventListener('click', _click,false);

		}



	},

	// первая точка на области примера
	first_coord: function () {

		this.main_area = this.example_area;

		// у примера нет смещения, т.к. нет курсора мышки
		this.otstup_x2[this.example_area] = 0;
		this.otstup_y2[this.example_area] = 0;

		// ставим точку на "примере".
		this.click({pageX:this.example_coords[0][0], pageY:this.example_coords[0][1]});
		this.main_area = this.main_area_;

		// возвращаем отступы
		var pos = this.getPos(this.main_area);
		this.otstup_x2[this.main_area] = pos.left;
		this.otstup_y2[this.main_area] = pos.top;

	},

	clear: function () {

		var drawingCanvas = document.getElementById( this.main_area );
		if (drawingCanvas && drawingCanvas.getContext) {
			var context = drawingCanvas.getContext('2d');
			context.clearRect(0, 0, 290, 414);
		}
		var drawingCanvas = document.getElementById( this.example_area );
		if (drawingCanvas && drawingCanvas.getContext) {
			var context = drawingCanvas.getContext('2d');
			context.clearRect(0, 0, 290, 414);
		}

		this.user_coords = [];
		this.arr_i = [];
		this.coords_array = [];
		this.otstup_x[this.main_area]=0;
		this.otstup_y[this.main_area]=0;



		// если это очищение области, где выводились точки из БД
		if (this.USER_COORDS) {
			this.USER_COORDS = false;
			coords.getObject(this.main_area).init();
		}
		else { // если это очищение области, где юзер только что ставил точки
			this.first_coord();
		}

	},

	mousemove:  function (ev) {

			x = ev.pageX-this.otstup_x2[ this.main_area ];
			y = ev.pageY-this.otstup_y2[ this.main_area ];


			// Y-axis
			var drawingCanvas = document.getElementById( this.for_mouse_move );
			if(drawingCanvas && drawingCanvas.getContext) {
				var context = drawingCanvas.getContext('2d');
				context.clearRect(0, 0, 290, 414);
				context.beginPath()
				context.moveTo(x,y);
				context.lineWidth=1;
				context.strokeStyle = "#000000";
				context.lineTo( x -  this.otstup_x[this.main_area] , y + 2048 );
				context.lineTo( x +  this.otstup_x[this.main_area] , y - 2048 );
				context.stroke();
			}

			// X-axis
			var drawingCanvas = document.getElementById( this.for_mouse_move );
			if(drawingCanvas && drawingCanvas.getContext) {
				var context = drawingCanvas.getContext('2d');
				context.beginPath()
				context.moveTo(x,y);
				context.lineWidth=1;
				context.strokeStyle = "#000000";
				context.lineTo( x + 2048 , y + this.otstup_y[this.main_area] );
				context.lineTo( x - 2048 , y - this.otstup_y[this.main_area] );
				context.stroke();
			}


	},

	click:  function (ev) {

		if (!this.arr_i[ this.main_area ])
			this.arr_i[ this.main_area ] = 0;

		if (!this.coords_array[ this.main_area ])
			this.coords_array[ this.main_area ] = [];

		if ( !this.coords_array[ this.main_area ][ this.arr_i[ this.main_area ] - 1 ] )
			this.coords_array[ this.main_area ][ this.arr_i[ this.main_area ] - 1 ] = [ ];

		// счетчик для массива координат. Для области с примером и для рабочей области
		this.arr_i[ this.main_area ] ++;

		// заполняем главный массив координат, по которому рисуем линии.
		x = ev.pageX - this.otstup_x2[ this.main_area ];
		y = ev.pageY - this.otstup_y2[ this.main_area ];
		this.coords_array[ this.main_area ][ this.arr_i[ this.main_area ] - 1 ] = Array(x, y);


    	// отмечаем точку
		this.number = this.arr_i[ this.main_area ];
		this.cross ( Array(x, y) );

        $('#comment-'+this.type).text( this.example_coords[ this.arr_i[ this.main_area ] - 1 ][ 2 ] );

		// выполняем действие, соответствующее номеру клика
		var action_arr = this.example_coords[ this.arr_i[ this.main_area ] - 1 ][ 3 ];
		if (action_arr) {

            for (i=2; i<action_arr.length; i++) {

                action = action_arr[i];

                switch ( action ) {

                   // задаем направление осей
                    case 'axes_direction':

                        x_katet_pril = (this.coords_array[this.main_area][action_arr[0]][0]-this.coords_array[this.main_area][action_arr[1]][0]);
                        y_katet_protiv = (this.coords_array[this.main_area][action_arr[0]][1]-this.coords_array[this.main_area][action_arr[1]][1]);
                        tang = y_katet_protiv/x_katet_pril;
                        atan = Math.atan(tang);
                        this.ugol[this.main_area] = atan*180/Math.PI;
                        this.otstup_x[this.main_area] = Math.tan((this.ugol[this.main_area])*Math.PI/180)*2048;
                        this.otstup_y[this.main_area] = Math.tan((this.ugol[this.main_area])*Math.PI/180)*2048;
                        //alert(this.otstup_x[this.main_area]+':'+this.otstup_y[this.main_area]);
                        break;

                    case 'x_line':

                        this.x_line(this.coords_array[this.main_area][action_arr[0]]);

                        break;

                    case 'y_line':

                        this.y_line(this.coords_array[this.main_area][action_arr[0]]);

                        break;

                    case 'draw_angle_left_bottom':

                        this.draw_angle(this.coords_array[this.main_area][action_arr[0]], this.coords_array[this.main_area][action_arr[1]], 'left_bottom',  this.ugol[this.main_area]);

                        break;

                    case 'draw_angle_right_bottom':

                        this.draw_angle(this.coords_array[this.main_area][action_arr[0]], this.coords_array[this.main_area][action_arr[1]], 'right_bottom',  this.ugol[this.main_area]);

                        break;

                    case 'p2p':

                        this.cnv_2_coord(this.coords_array[this.main_area][action_arr[0]], this.coords_array[this.main_area][action_arr[1]], false, '#ff0000');

                        break;

                    // линия по оси y от центра точек
                    case 'y_center':

                        var center_point = [];
                        center_point['x'] = this.coords_array[this.main_area][action_arr[0]][0]+((this.coords_array[this.main_area][action_arr[1]][0]-this.coords_array[this.main_area][action_arr[0]][0])/2);
                        center_point['y'] = this.coords_array[this.main_area][action_arr[0]][1]+((this.coords_array[this.main_area][action_arr[1]][1]-this.coords_array[this.main_area][action_arr[0]][1])/2);
                        this.cnv_2_coord(Array(center_point['x'], center_point['y']), Array(center_point['x']-this.otstup_x[this.main_area], center_point['y']+2048), false, this.line_color);
                        this.cnv_2_coord(Array(center_point['x'], center_point['y']), Array(center_point['x']+this.otstup_x[this.main_area], center_point['y']-2048), false, this.line_color);

                        break;

                }
            }
        }

		//alert( this.coords_array[ this.main_area ] );

		// при клике по рабочей области, делаем клик в область примера. Только если это не авто-заполнение
		if ( this.main_area != this.example_area && !this.USER_COORDS) {

            console.log( this.main_area +' /  '+this.example_area+' / '+ this.ugol[this.main_area]);
			
			// если это последний клик, то шлем все точки в БД			
			if ( this.example_coords.length == this.arr_i[ this.main_area ] ) {

				var json = JSON.stringify(this.coords_array[this.main_area]);

				$.post( 'ajax?controllerName=saveUserCoords',
                    {'type' : this.main_area, 'coords_json' : json },
                    function (data) {
                        alert('Saved');
                    }
                );

				
			}
			else {
		
				this.main_area=this.example_area;
				
				this.click( {
						pageX: this.example_coords[ this.arr_i[ this.example_area ] ][ 0 ], 
						pageY: this.example_coords[ this.arr_i[ this.example_area ] ][ 1 ]
					} );
			}
		} 
		else if ( !this.USER_COORDS ) {
			
			//alert(1);
			
			this.main_area=this.main_area_;
			var pos = this.getPos(this.main_area);
			this.otstup_x2[this.main_area] = pos.left;
			this.otstup_y2[this.main_area] = pos.top;
		}

		//alert(2);
		
		
		
	}
}
