// Glowing animation
var glow = $('.glowing');
setInterval(function () {
  glow.toggleClass('glow');
}, 1000);


/*
////////////////////////////
////////////////////////////
  Connect to websocket
///////////////////////////
///////////////////////////
*/

const hostUrl = "wss://api.jmatray.me/v1/ws?auth="
const auth = "placeholder"

socket = new WebSocket(hostUrl + auth)
socket.onopen = function() {
    alert("WebSocket Opened");
}

socket.onmessage = function(event) {
  console.log("Message Received");

  var receivedMsg = JSON.parse(event.data);
  var type = receivedMsg.type;

  if (type === 'new-lobby') {

  } 
}

socket.onclose = function(event) {
  //window.location.replace("http://127.0.0.1:8080/clients/trivia/app/public/sign_in.html");
}

socket.onerror = function(event) {
  //alert("Please Sign-In")
  //window.location.replace("http://127.0.0.1:8080/clients/trivia/app/public/sign_in.html");
}







/*
////////////////////////////
////////////////////////////
  DOM Functions
///////////////////////////
///////////////////////////
*/

// Switch from landing page to lobby
function switchToLobby() {
  var landing = document.querySelector(".landing");
  var game = document.querySelector(".game");
  if (landing.style.display === "none") {
    landing.style.display = "flex";
    game.style.display = "none"
  } else {
    landing.style.display = "none";
    game.style.display = "flex"
  }
};

// Switch from lobby to game
function switchToGame() {
  $('.waiting').hide();
  $('.playing').show();
  $('.board').css('height', 'auto')
}



// Switch from game to landing
function leaveGameHandler() {
  $('.game').hide();
  $('.waiting').show();
  $('.playing').hide();
  $('.landing').show();
  $('.board').css('height', '80vh');
}



/*
////////////////////////////
////////////////////////////
  Landing Page Functions
///////////////////////////
///////////////////////////
*/

function createAddLobby (lob) {
  var newLob = document.createElement('DIV');
  newLob.setAttribute('class', 'lobby')

  var img = document.createElement('IMG')
  img.setAttribute('class', 'lobby-pic')
  img.setAttribute('src', '/imgs/Drawing.png')

  var creat = document.createElement('P')
  creat.innerHTML = "Creator: " + lob.creator

  var cat = document.createElement('P')
  cat.innerHTML = "Category: " + lob.category

  var diff = document.createElement('P')
  diff.innerHTML = "Difficulty: " + lob.difficulty

  var join = document.createElement('button')
  join.setAttribute('id', lob.id)
  join.setAttribute('class', 'join button')
  if (lob.inProgress === false) {
    $(join).text("Join")
    $(join).on('click', switchToLobby)
  } else {
    $(join).text("In Progress").addClass("disabled");

  }

  newLob.appendChild(img)
  newLob.appendChild(creat)
  newLob.appendChild(cat)
  newLob.appendChild(diff)
  newLob.appendChild(join)
  $('.lobbies').append(newLob);
}

function joinGame() {
    /*
    Step 1: Post request to /trivia/id
    Step 2: Wait for response with lobby struct
    Step 3: Update number of players
    Step 4: Switch to show lobby
  */
}

function createGame () {
  /*
    Step 1: Post request to /trivia
    Step 2: Wait for response of new lobby created, with lobby struct
    Step 3: Track the creator somehow
    Step 4: Update number of players
    Step 4: Switch to show lobby
  */

  switchToLobby()
}
$('.new-lobby').on('click', createGame);




// Get lobbies
let placeholderLobs = [{ id: "1", creator: "Dalai", category: 'Nature', difficulty: 'Easy', inProgress: false }]
placeholderLobs.forEach(lob => createAddLobby(lob))

function getAllLobbies() {
  /*
    Step 1: remove all elements from lobbies DOM
    Step 2: send get request to get all lobby structs
    Step 3: Wait for response with all lobbies
    Step 4: Loop through each lobby, adding to DOM
  */
}

var lobbies = document.querySelectorAll(".join");
var gameStart = document.querySelectorAll(".start-game");
var leaveGame = document.querySelector(".leave-game");
leaveGame.addEventListener('click', leaveGameHandler, false);
for (var i = 0; i < gameStart.length; i++) {
  gameStart[i].addEventListener('click', startGameHandler, false)
}


/*
////////////////////////////
////////////////////////////
  Inside Lobby Functions
///////////////////////////
///////////////////////////
*/
$(".form-control").on('change', function () {
  console.log(this.value)
});


// Handle the start of a new game
function startGameHandler() {
  /*
    Step 1: Run newQuestionHandler
    Step 2: Switch to show game DOM
    Step 3: (Might have to wait for question)
  */

 // $('#' + id).addClass("disabled").text("In Progress").off('click', switchToLobby);
  newQuestionHandler()
  switchToGame()
}






/*
////////////////////////////
////////////////////////////
  Inside Game Functions
///////////////////////////
///////////////////////////
*/

// Handle new question messages from server

function newQuestionHandler() {
  // TEMPORARY TIME FOR QUESTION
  var now = new Date().getTime()
  var timeLeft = 30

  clearInterval(timerId);
  $('.timer').html('');
  $('#ans1').html('Test');

  var timerId = setInterval(countdown, 1000);

  function countdown() {
    $('.timer').html(timeLeft + 's')
    if (timeLeft == 0) {
      clearTimeout(timerId);
      doSomething();
    } else {
      timeLeft--;
    }
  }
}

function submitAnswer() {
  /*
    Step 1: Send post request to answer
    Step 2: After successful post request, show message that answer was submitted
    Step 3: Disable answer buttons
  */
}


















