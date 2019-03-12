// Glowing animation
var glow = $('.glowing');
setInterval(function () {
  glow.toggleClass('glow');
}, 1000);



// Helper Functions
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



// Button Handlers
var currentLobby;
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

function joinGame() {
    /*
    Step 1: Post request to /trivia/id
    Step 2: Wait for response with lobby struct
    Step 3: Switch to show lobby
  */
}

function createGame () {
  /*
    Step 1: Post request to /trivia
    Step 2: Wait for response of new lobby created, with lobby struct
    Step 3: Track the creator somehow
    Step 4: Switch to show lobby
  */

  switchToLobby()
}
$('.new-lobby').on('click', createGame);

$(".form-control").on('change', function () {
  console.log(this.value)
});

$('#category').val('lmao')




/*
/////
  Websocket Message Handlers
/////
*/

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










/*
/////
  Game Logic
/////
*/

// Handle new question messages from server

newQuestionHandler = function () {
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

submitAnswer = function() {

}

// Switch DOM to show the 'playing' div, and run 'startGameHandler'
var switchToGame = function () {
  startGameHandler(currentLobby);
  newQuestionHandler()
  $('.waiting').hide();
  $('.playing').show()
  $('.board').css('height', 'auto')
}

// Handle the start of a new game
startGameHandler = function (id) {
  $('#' + id).addClass("disabled").text("In Progress").off('click', switchToLobby);
}

// Handles switching back to the landing page
var leaveGameHandler = function () {
  $('.game').hide();
  $('.waiting').show();
  $('.playing').hide();
  $('.landing').show();
  $('.board').css('height', '80vh');
}


var lobbies = document.querySelectorAll(".join");
var gameStart = document.querySelectorAll(".start-game");
var leaveGame = document.querySelector(".leave-game");
leaveGame.addEventListener('click', leaveGameHandler, false);
for (var i = 0; i < gameStart.length; i++) {
  gameStart[i].addEventListener('click', switchToGame, false)
}









