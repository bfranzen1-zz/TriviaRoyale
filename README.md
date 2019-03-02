# TriviaRoulette
_03.01.2019_

Jake Matray

Blake Franzen

## Overview
Trivia Roulette is an application that lets users play trivia games with others in a format similar to popular battle royale games. 

## Who?
The specific target audience for this project are those that heavily enjoy trivia games, who are also looking to both show off and advance their knowledge in specific areas. We envision that those who want to put their knowledge to the test by competing against others in a fast-paced and challenging setting would be the ones most likely to use our application.

## Why?
Despite the massive player base of Battle Royale video games, they are still missing the market of those that simply do not enjoy video games. Because the desire to compete and receive the satisfaction of being number one out of a large group of people is not solely found in people who enjoy video games, people would want to use our application to get that outlet they can’t find elsewhere. If you are someone who enjoys trivia, whether it be by watching game shows or partaking in it themselves, or even just want to be able to show off how smart you are, you would be given the opportunity to measure your skills against others while also improving your own knowledge. 
As developers, this application is not only an opportunity to show off our skill set learned during the course but to also push the boundaries of our understanding of concepts. We hope that it will also introduce us to new technologies and services to broaden the scope of our knowledge as well. Most of all, we wanted to build a fun and engaging application to share with others. 

## Architecture

![alt text](images/image1.jpg)

## User Stories

| Priority | User                 | Description                                                                                   | Implementation                                                                                                                                                                                                                                                 |
|----------|----------------------|-----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| P0       | As a user            | I want to create a new account with Trivia Roulette to play with others and track my progress | Upon receiving a  **POST request to /v1/users** , UsersHandler will validate and insert a new user into the user store. UsersHandler will then begin a session for the new user in the session store                                                               |
| P1       | As a registered user | I want to join a game and play against others players                                         | Upon recieivng a  **POST request to /v1/trivia/{triviaID}**,  the trivia microservice adds the authenticated user to the state of the desired game lobby                                                                                                           |
| P1       | As a group of users  | I want to submit an answer to a question within a game                                        | After receiving a  **GET request to /v1/trivia/answer**,  containing a json list of users and their answers,   the trivia microservice will determine which of the answers are correct, and respond with the list of users and whether or not they were correct.   |
| P2       | As a registered user | I want my statistics to be recorded after each game                                           | Upon receiving a  **POST request to /v1/users/{userID}**,  with a json object containing the number of correct answers, whether or not they won, and how many points they received for the game, and StatsHandler will insert that information into UserStatistics |
| P3       | As a registered user | I want to view my statistics, or another player’s statistics                                  | Upon receiving a  **GET request to /v1/users/{userID}**,  StatsHandler will query the UserStatistics table for each with provided userID, and return a json object containing the sum for each statistic category                                                  |
| P2       | As a registered user | I want to send chat messages to other players in my game                                      | Upon receiving a  **POST request to /v1/trivia/{triviaID}/messages**,  the messaging microservice will insert the message body and the associated triviaID into the Message table                                                                                  |
| P2       | As a registered user | I want to view chat messages sent by other players in my game                                 | Upon receiving a  GET request to **/v1/trivia/{triviaID}/messages**,  the messaging microservice will respond with a list of all the messages for that game                                                                                                        |


## Schemas

## Endpoints 
The trivia microservice will handle the state of the game. The state will include a list of questions to ask and the users currently in the game. The service will generate a set of questions to ask users based on the API we are pulling from. After each question is answered by users in the alloted time those that didn’t answer correctly will be removed from the game. 

POST /v1/users
- Given a json object containing user information, including email, password, first name, last name, user name, and a photoURL, inserts the user into the user store
- Returns the newly inserted user, along with their ID
- Creates a session for the newly inserted user in the session store


GET /v1/users/{userID}
- Given a userID, will return the user name and statistics for that specific userID 

POST /v1/user/{userID}
- Adds user game statistics to UserStatistics table after a game has been completed

POST /v1/trivia
- Insert a new trivia game into the trivia microservice

POST /v1/trivia/{triviaID}
- Add user to lobby
- Upgrade to websocket to send questions
- Start game if time runs out on lobby waiting

GET /v1/trivia/{triviaID}/answer
- User sends answer to question
- Game state updated
- If answer incorrect user removed from state
- Send user result of answer

POST /v1/trivia/{triviaID}/messages
- Inserts the message body and the creator of the message into the messaging microservice

GET /v1/trivia/{triviaID}/message
- Returns all messages for the request trivia game, along with their creators

POST /v1/sessions
- Validate and insert the provided user into the session store

DELETE /v1/sessions
- Remove the provided user from the session store


We will also utilize the API endpoints we defined in assignments throughout the quarter for signing users up, storing session information for active users, and sending/receiving messages. 

