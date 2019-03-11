// Glowing animation
var glow = $('.glowing');
setInterval(function () {
  glow.toggleClass('glow');
}, 1000);



// Helper Functions
var createAddLobby = function (lob) {
  var newLob = document.createElement('DIV');
  newLob.setAttribute('class', 'lobby')
  
  var img = document.createElement('IMG')
  img.setAttribute('class', 'lobby-pic')
  img.setAttribute('src', 'imgs/Drawing.png')

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
var switchToLobby = function  () {
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

var createGame = function () {
  switchToLobby()
}
$('.new-lobby').on('click', createGame);

$( ".form-control" ).on('change', function() {
  console.log(this.value)
});

$('#category').val('lmao')

/*
/////
  Websocket Message Handlers
/////
*/

// Get lobbies




let placeholderLobs = [{id: "1",creator: "Dalai", category: 'Nature', difficulty: 'Easy', inProgress: false}]

//var getAllLobbies = function() {
  placeholderLobs.forEach(lob => createAddLobby(lob))
//}





var lobbies = document.querySelectorAll(".join");

var gameStart = document.querySelectorAll(".start-game");
var leaveGame = document.querySelector(".leave-game");




var switchToGame = function () {
  startGameHandler(currentLobby);
  $('.waiting').hide();
  $('.playing').show()
  $('.board').css('height', 'auto')
}

startGameHandler = function (id) {
  $('#' + id).addClass("disabled").text("In Progress").off('click', switchToLobby);
}

var leaveGameHandler = function () {

  $('.game').hide();
  $('.waiting').show();
  $('.playing').hide();
  $('.landing').show();
  $('.board').css('height', '80vh');
}

leaveGame.addEventListener('click', leaveGameHandler, false);

for (var i = 0; i < gameStart.length; i++) {
  gameStart[i].addEventListener('click', switchToGame, false)
}






