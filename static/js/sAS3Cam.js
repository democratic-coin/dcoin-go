
			var jcrop_api = '';
function webcam1() {
    $("#webcam").webcam({
        swffile: "static/swf/sAS3Cam.swf?v="+Math.random(),

        previewWidth: 640,
        previewHeight: 480,

        resolutionWidth: 640,
        resolutionHeight: 480,

        /**
         * Determine if we need to stretch or scale the captured stream
         *
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/Stage.html#scaleMode
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/StageScaleMode.html
         */
        StageScaleMode: 'noScale', //

        /**
         * Aligns video output on stage
         *
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/StageAlign.html
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/Stage.html#align
         * Empty value defaults to "centered" option
         */
        StageAlign: 'TL',

        noCameraFound: function () {
            this.debug('error', 'Web camera is not available');
        },

        swfApiFail: function(e) {
            this.debug('error', 'Internal camera plugin error');
        },

        cameraDisabled: function () {
            this.debug('error', 'Please allow access to your camera');
        },

        debug: function(type, string) {
            console.log(string);
            if (type == 'error') {
                //$(".webcam-error").html(string);
            } else {
                //$(".webcam-error").html('');
            }
        },

        cameraEnabled:  function () {
            this.debug('notice', 'Camera enabled');
            var cameraApi = this;
            if (cameraApi.isCameraEnabled) {
                return;
            } else {
                cameraApi.isCameraEnabled = true;
            }
            var cams = cameraApi.getCameraList();

            for(var i in cams) {
                $("#popup-webcam-cams").append("<option value='"+i+"'>" + cams[i] + "</option>");
            }

            setTimeout(function() {
                $("#popup-webcam-take-photo").removeAttr('disabled');
                $("#popup-webcam-take-photo").show();
                cameraApi.setCamera('0');
            }, 750);

            $("#popup-webcam-cams").change(function() {
                var success = cameraApi.setCamera($(this).val());
                if (!success) {
                    cameraApi.debug('error', 'Unable to select camera');
                } else {
                    cameraApi.debug('notice', 'Camera changed');
                }
            });

            $('#popup-webcam-take-photo').click(function() {
                var result = cameraApi.save();
                console.log(result);
                if (result && result.length) {
                    var actualShotResolution = cameraApi.getResolution();
                    var image = new Image();
                    image.src = 'data:image/jpeg;base64,' + result;
                    console.log(result);
                    image.onload = function() {
                        var photo_type = $('#photo_type').val();
                        $('#' + photo_type + '_photo').attr('width', 350);
                        var k = this.width / 350;
                        var new_height = Math.round(this.height / k);
                        $('#' + photo_type + '_photo').attr('height', new_height);
                        $('#' + photo_type + '_photo_div').css('width', 350);

                        var c = document.getElementById(photo_type + "_photo");
                        var ctx = c.getContext("2d");
                        ctx.drawImage(image, 0, 0, this.width, this.height, 0, 0, 350, new_height);
                        if (first_load == false) {
                            $('#' + photo_type + '_photo').cropper("destroy");
                        }
                        crop_img('#' + photo_type + '_photo');
                        first_load = false;
                    }

                   // $('#img_b64').val(result);
                    //$("#result").html('<img src="data:image/jpeg;base64,' + result + '" id="result_img"><p>'+window['crop_img_text']+'</p><button  id="save"  type="button" class="btn btn-default" style="margin-left: 320px" onclick=\'save_crop()\'>Save</button>');

                    /*$('#result_img').Jcrop({
                        onChange: showCoords,
                        onSelect: showCoords,
                        bgColor: 'black',
                        bgOpacity: .4,
                        aspectRatio:7/10
                    });
                    jcrop_api = $('#result_img').data('Jcrop');*/

                    // alert('base64encoded jpeg (' + actualShotResolution[0] + 'x' + actualShotResolution[1] + '): ' + result.length + 'chars');

                    /* resume camera capture */
                    cameraApi.setCamera($("#popup-webcam-cams").val());
                    console.log('cameraApi.setCamera($("#popup-webcam-cams").val())');
                    webcam1();
                } else {
                    cameraApi.debug('error', 'Broken camera');
                }
            });


            var reload = function() {
                $('#popup-webcam-take-photo').show();
            };

            $('#popup-webcam-save').click(function() {
                reload();
            });
        }
    });
}
			$(document).ready(function() {

                webcam1();

			});

/*
            function send1 (file_id, progress, img_id, type, crop_img_text, save) {
                var
                    $f = $('#'+file_id),
                    $p = $('#'+progress),
                    up = new uploader($f.get(0), {
                        url:'ajax/upload.php',
                        prefix:'image',
                        type:type,
                        progress:function(ev){ $p.html(((ev.loaded/ev.total)*100)+'%'); $p.css('width',$p.html()); },
                        error:function(ev){
                            alert('error');
                        },
                        success:function(data){

                            if (data.error !== undefined) {
                                alert(data.error)
                            }
                            else {
                                $p.html('100%');
                                $p.css('width',$p.html());
                                $('#'+img_id).html(crop_img_text+'<br><img width="350" src="'+data.url+'?r='+Math.random()+'" id="'+type+'"><p><button onclick="send_crop(\''+type+'\', \'coords\', \''+img_id+'\')"  type="button" class="btn btn-default">'+save+'</button></p>');
                                crop_img ();
                                $('#'+type).Jcrop({
                                      onSelect: showCoords,
                                      onChange: showCoords,
                                      bgColor:     'black',
                                      bgOpacity:   .4,
                                      aspectRatio: 7/10
                                 });
                            }
                        }
                    });

                up.send();

            }*/

            $( "#from_webcam_show" ).click(function(e) {
                console.log('from_webcam_show');
                $("#from_webcam").css("display", "block");
                $("#from_file_form").css("display", "none");
                e.preventDefault();
                e.stopPropagation();
            });
            $( "#from_file_show" ).click(function(e) {
                console.log('from_file_show');
                $("#from_file_form").css("display", "block");
                $("#from_webcam").css("display", "none");
                e.preventDefault();
                e.stopPropagation();
            });

            function showCoords(c) {
                $('#coords').text(c.x+';'+c.y+';'+c.x2+';'+c.y2+';'+c.w+';'+c.h);
            };

            function save_crop () {
                console.log('save');
                var img_type = $('#img_type').val();
                $.post('ajax/crop_photo.php', {'coords': $('#coords').text(), 'image': $('#img_b64').val(), 'type': 'user_'+img_type+'_webcam' },
                    function(data) {
                        // jcrop_api.destroy();
                        $('#result0').html('<table><tr><td><img src="img/'+img_type+'.jpg"></td><td><div id="result" style="width:640px; height:480px;"><img width="350" src="'+data.url+'?r='+Math.random()+'"></div></td></tr></table>');
                    }
                    , 'json');
            }
