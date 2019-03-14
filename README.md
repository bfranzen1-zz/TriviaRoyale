# TriviaRoyale

_03.01.2019_

Jake Matray

Blake Franzen

## Overview
TriviaRoyale is an application that lets users play trivia games with others in a format similar to popular [battle royale](https://en.wikipedia.org/wiki/Battle_royal) games. 

## Who?
The specific target audience for this project are those that heavily enjoy trivia games, who are also looking to both show off and advance their knowledge in specific areas. We envision that those who want to put their knowledge to the test by competing against others in a fast-paced and challenging setting would be the ones most likely to use our application.

## Why?
Despite the massive player base of Battle Royale video games, they are still missing the market of those that simply do not enjoy video games. Because the desire to compete and receive the satisfaction of being number one out of a large group of people is not solely found in people who enjoy video games, people would want to use our application to get that outlet they can’t find elsewhere. If you are someone who enjoys trivia, whether it be by watching game shows or partaking in it themselves, or even just want to be able to show off how smart you are, you would be given the opportunity to measure your skills against others while also improving your own knowledge. 
As developers, this application is not only an opportunity to show off our skill set learned during the course but to also push the boundaries of our understanding of concepts. We hope that it will also introduce us to new technologies and services to broaden the scope of our knowledge. Most of all, we wanted to build a fun and engaging application to share with others. 

## Architecture

![architecture diagram](images/image1.jpg)

## User Stories

| Priority | User                 | Description                                                                                   | Implementation                                                                                                                                                                                                                                                 |
|----------|----------------------|-----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| P0       | As a user            | I want to create a new account with Trivia Royale to play with others and track my progress | Upon receiving a  **POST request to /v1/users** , UsersHandler will validate and insert a new user into the user store. UsersHandler will then begin a session for the new user in the session store                                                               |
| P1       | As a registered user | I want to join a game and play against others players                                         | Upon recieivng a  **POST request to /v1/trivia/{triviaID}?type=add**,  the trivia microservice adds the authenticated user to the state of the desired game lobby                                                                                                           |
| P1       | As a user | I want to submit an answer to a question within a game                                        | After receiving a  **POST request to /v1/trivia/{triviaID}?type=answer**,  containing a json object with the users answer,   the trivia microservice will add that to the list of answers for a specific question in the game, for later evaluation   |
| P2       | As a registered user | I want to access my statistics                                          | Upon calling the  **GET request to /v1/trivia/user/{userID}**, the user will get a json object containing the games played, points received, and number of wins |                                                 |
| P2       | As a registered user | I want to send chat messages to other players                                     | Upon receiving a  **POST request to /v1/channels/{channelID}**,  the messaging microservice will insert the message body into the general chat bar in the trivia microservice                                                                                  |
| P2       | As a registered user | I want to view chat messages sent by other players in my game                                 | When a message is posted to the message microservice it will post the message to RabbitMQ and be displayed on the trivia microservice                                                                                                   |


## Schemas

![database structure](images/image4.jpg)

## Endpoints 
The trivia microservice will handle the state of the game. The state will include a list of questions to ask and the users currently in the game. The service will generate a set of questions to ask users based on the API we are pulling from. After each question is answered by users in the alloted time those that didn’t answer correctly, or in time, will be removed from the game. 

GAME LOGIC
- On game start, a go routine is initiatied that sends questions periodically depending on the difficulty set for the game. 
    - the questions are sent to the RabbitMQ service which posts "game-question" messages containing a question struct to all users that are players in the lobby 
- After a period of waiting the server checks that answers provided by users and removes those that didn't answer in time or got the wrong answer. 
- The game has ended when there is either one user left or there are no more questions to answer. When the game ends or a user loses they are kicked from the lobby and their statistics for the game are saved. 
    - a message is sent to the Rabbit queue "game-over" or "game-won" and contains the lobby state at that point in time
- Points are awarded for correct answers and the situation in which someone won (number of players left and difficulty)

POST /v1/users
- Given a json object containing user information, including email, password, first name, last name, and user name, inserts the user into the user store
- Returns the newly inserted user, along with their ID
- Creates a session for the newly inserted user in the session store
- Requires a JSON object with:
    - string first name of user
    - string last name of user 
    - string email of user
    - string password of user
    - string passwordconf of user
- 400 errors are returned if invalid data is sent
- 201 is returned on success

TRIVIA MICROSERVICE 
- 401 errors are returned if at any point a user is not authenticated 


GET /v1/trivia
- If the user is authenticated it gets a json object containing an array of active lobbies
- 401 error returned if user not authenticated
- 200 on success

POST /v1/trivia
- Insert a new trivia game into the trivia microservice with the passed options encoded in a json object if the user is authenticated
- Requires JSON object with:
    - an int representing the number of questions 
    - an int representing the max players to allow
    - an int representing the category to select
    - a string representing the difficulty (easy, medium, hard)
- 400 is returned if data is invalid
- a 200 is returned on success 
    - additionally, a message is sent to the RabbitMQ type "lobby-new" containing the new lobby struct

GET /v1/trivia/{triviaID}
- the authenticated creator of the lobby wants to start the game
- go routine is initiated that starts game and handles state
- a 200 status is returned if the start succeeded 

POST /v1/trivia/{triviaID}?type=
- checks if user is authenticated
- Add user to lobby
- Upgrade to websocket to send questions
- Start game if the max players is reached or the creator starts the game
- if the type query is set to add it will add the user to the lobby, if it's answer it will submit the answer to the current question in the lobby
- if the added user makes the lobby full the game is started
- a 201 status is returned on success
    - additionally, a message is sent to the RabbitMQ type "lobby-add" containing the new lobby struct and userIds 

PATCH /v1/trivia/{triviaID}
- checks if user is the creator of the lobby and is authenticated
- changes the settings of the lobbies using the passed options encoded as json
- Requires JSON object with:
    - an int representing the number of questions 
    - an int representing the max players to allow
    - an int representing the category to select
    - a string representing the difficulty (easy, medium, hard)
- a 400 is returned if the data is invalid
- a 200 is returned on success
    - additionally, a message is sent to the RabbitMQ type "lobby-update" containing the new lobby struct

We will also utilize the API endpoints we defined in assignments throughout the quarter for signing users up, storing session information for active users, and sending/receiving messages via websockets. 

