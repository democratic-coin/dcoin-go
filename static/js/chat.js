		var lang = 1;
		var room = 0;
		var chatMessage;
		var decryptChatMessage;
		var chatEncrypted = 0;

		$('#sendToChat').bind('click', function () {

			/*document.getElementById("sendToChat").innerHTML="wait...";
			document.getElementById("sendToChat").disabled="true";*/
			$('#sendToChat').html('<img src="/static/img/squares.gif" style="width:20px; margin:0px">');
			$('#sendToChat').prop('disabled', true);

			setTimeout(function() {
				chatMessage = $('#myChatMessage').val();
				if (chatEncrypted == 1) {
					$.post('ajax?controllerName=encryptChatMessage', {
						'receiver': $('#chatUserIdReceiver').val(),
						'message': $('#myChatMessage').val()
					}, function (data) {
						chatMessage = data.success;
						decryptChatMessage = $('#myChatMessage').val();
						sendToTheChat()
					}, 'JSON');
				} else {
					sendToTheChat()
				}
			}, 500);

		});

		if (!Date.now) {
			Date.now = function() { return new Date().getTime(); }
		}
		function sendToTheChat() {


			var chatMessageReceiver =  $('#chatUserIdReceiver').val();
			var chatMessageSender =  userId;
			var status = 0;
			if (chatEncrypted == 1) {
				status = 1
			}
			var signTime = Math.floor(Date.now() / 1000);




			var objEx =
			{
				key: $("#key").text(),
				pass: $("#password").text(),
				forsign: lang+","+room+","+chatMessageReceiver+","+chatMessageSender+","+status+","+chatMessage+","+signTime,
			};

			var workerAjax = new Worker("/static/js/worker.js");
			workerAjax.onmessage  = function(event) {
				if (typeof event.data.error != "undefined") {
					$("#chat_alert").html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>'+event.data.error+'</p></div>');
					$('#sendToChat').prop('disabled', false);
					$('#sendToChat').html('Send');
				} else {
					$.post( 'ajax?controllerName=sendToTheChat', {
						'receiver': chatMessageReceiver,
						'sender': userId,
						'lang': lang,
						'room': room,
						'message': chatMessage,
						'decrypt_message': decryptChatMessage,
						'status': status,
						'sign_time': signTime,
						'signature': event.data.hSig
					}, function (data) {
						$('#sendToChat').prop('disabled', false);
						$('#sendToChat').html('Send');
					});
				}
			};
			workerAjax.onerror = function(err) {
				alert(err.message);
			};
			workerAjax.postMessage(objEx);

			//var e_n_sign = get_e_n_sign( $("#key").text(), $("#password").text(), lang+","+room+","+chatMessageReceiver+","+chatMessageSender+","+status+","+chatMessage+","+signTime, 'chat_alert');

		}

		function scrollToBottom() {
			var objDiv = document.getElementById("chatwindow");
			//console.log(objDiv.scrollHeight-67-objDiv.scrollTop)
			if (objDiv.scrollTop == 0 || objDiv.scrollHeight-67-objDiv.scrollTop == objDiv.clientHeight) {
				objDiv.scrollTop = objDiv.scrollHeight;
			}
		}
		$(document).ready(function() {

			$.post('ajax?controllerName=getChatMessages&first=1&room='+room+'&lang='+lang, {}, function (data) {
				if (typeof data.messages != "undefined" && data.messages != "") {
					$('#chatMessages').append(data.messages);
					scrollToBottom();
				}
					setTimeout(function() {
						var intervalID = setInterval( function() {

								/*$.ajax({
									url: 'ajax?controllerName=getChatMessages&room='+room+'&lang='+lang,
									type: 'POST',
									async: false,
									cache: false,
									dataType: "JSON",
									timeout: 3000,
									error: function(){
										return true;
									},
									success: function(data){
										$('#chatMessages').append(data.messages);
										scrollToBottom();
									}
								});*/

							$.post( 'ajax?controllerName=getChatMessages&room='+room+'&lang='+lang, {}, function (data) {
								//if(typeof data.messages != "undefined" && data.messages !="") {
								//console.log("data.messages", data.messages);
								$('#chatMessages').append(data.messages);
								scrollToBottom();
								if (data.chatStatus == "bad") {
									$('#chatTitle').html("Chat <span style='color:#EA6153'><i class='fa fa-power-off'></i></span>")
								} else {
									$('#chatTitle').html("Chat <span style='color:#37BC9B'><i class='fa fa-power-off'></i></span>")
								}
								//}
							}, 'JSON');

							var objDiv = document.getElementById("chatwindow");
							//console.log(objDiv.scrollHeight, objDiv.scrollTop, objDiv.clientHeight)

						} , 1000);
						intervalIdArray.push(intervalID);
					}, 500);


			}, 'JSON');


		});

		function setReceiver(nick, receiverId){
			$('#myChatMessage').val(nick+", ");
			$('#chatUserIdReceiver').val(receiverId);
			$("#selectReceiver").css("display", "none");
			$("#myChatMessage").css("display", "inline-block");
			console.log("receiverId", receiverId)
		}



		function lock_unlock() {
			if ($('#chatLockIco').attr('class') == "fa fa-lock") {
				$('#chatLockIco').attr("class", "fa fa-unlock");
				$("#myChatMessage, #sendToChat").css("display", "block");
				$("#selectReceiver").css("display", "none");
				$("#myChatMessage").css("background-color", "#fff");
				$("#myChatMessage").css("color", "#000");
				chatEncrypted = 0
			} else {
				$('#chatLockIco').attr("class", "fa fa-lock");
				if ($("#chatUserIdReceiver").val() == "0") {
					$("#myChatMessage, #sendToChat").css("display", "none");
					$("#selectReceiver").css("display", "inline-block");
				}
				$("#myChatMessage").css("background-color", "#BC5247");
				$("#myChatMessage").css("color", "#fff");
				chatEncrypted = 1
			}
		}

