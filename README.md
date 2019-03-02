# Trivia Roulette
<p style="color: red; font-weight: bold">>>>>>  gd2md-html alert:  ERRORs: 0; WARNINGs: 1; ALERTS: 3.</p>
<ul style="color: red; font-weight: bold"><li>See top comment block for details on ERRORs and WARNINGs. <li>In the converted Markdown or HTML, search for inline alerts that start with >>>>>  gd2md-html alert:  for specific instances that need correction.</ul>

<p style="color: red; font-weight: bold">Links to alert messages:</p><a href="#gdcalert1">alert1</a>
<a href="#gdcalert2">alert2</a>
<a href="#gdcalert3">alert3</a>

<p style="color: red; font-weight: bold">>>>>> PLEASE check and correct alert issues and delete this message and the inline alerts.<hr></p>




<p id="gdcalert1" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Project-proposal0.png). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert2">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Project-proposal0.png "image_tooltip")



# Trivia Roulette


## 02.27.2019

**─**

Jake Matray

Blake Franzen


# Overview

Trivia Roulette is an application that lets users play trivia games with others in a format similar to popular battle royale games. 


# Who?

	The specific target audience for this project are those that heavily enjoy trivia games, who are also looking to both show off and advance their knowledge in specific areas. We envision that those who want to put their knowledge to the test by competing against others in a fast-paced and challenging setting would be the ones most likely to use our application.


# Why?

Despite the massive player base of Battle Royale video games, they are still missing the market of those that simply do not enjoy video games. Because the desire to compete and receive the satisfaction of being number one out of a large group of people is not solely found in people who enjoy video games, people would want to use our application to get that outlet they can't find elsewhere. If you are someone who enjoys trivia, whether it be by watching game shows or partaking in it themselves, or even just want to be able to show off how smart you are, you would be given the opportunity to measure your skills against others while also improving your own knowledge. 

As developers, this application is not only an opportunity to show off our skill set learned during the course but to also push the boundaries of our understanding of concepts. We hope that it will also introduce us to new technologies and services to broaden the scope of our knowledge as well. Most of all, we wanted to build a fun and engaging application to share with others. 


# Architecture


## 

<p id="gdcalert2" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Project-proposal1.jpg). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert3">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Project-proposal1.jpg "image_tooltip")



# User Stories


<table>
  <tr>
   <td>Priority
   </td>
   <td>User
   </td>
   <td>Description
   </td>
   <td>Implementation
   </td>
  </tr>
  <tr>
   <td><strong>P0</strong>
   </td>
   <td><strong>As a user</strong>
   </td>
   <td><strong>I want to create a new account with Trivia Roulette to play with others and track my progress </strong>
   </td>
   <td><strong>Upon receiving a </strong>POST request to /v1/users<strong>, UsersHandler will validate and insert a new user into the user store. UsersHandler will then begin a session for the new user in the session store</strong>
   </td>
  </tr>
  <tr>
   <td><strong>P1</strong>
   </td>
   <td><strong>As a registered user</strong>
   </td>
   <td><strong>I want to join a game and play against others players</strong>
   </td>
   <td><strong>Upon recieivng a </strong>POST request to /v1/trivia/{triviaID}, <strong>the trivia microservice adds the authenticated user to the state of the desired game lobby</strong>
   </td>
  </tr>
  <tr>
   <td><strong>P1</strong>
   </td>
   <td><strong>As a group of users</strong>
   </td>
   <td><strong>I want to submit an answer to a question within a game</strong>
   </td>
   <td><strong>After receiving a </strong>GET request to /v1/trivia/answer, <strong>containing a json list of users and their answers,</strong> <strong>the trivia microservice will determine which of the answers are correct, and respond with the list of users and whether or not they were correct.</strong>
   </td>
  </tr>
  <tr>
   <td><strong>P2</strong>
   </td>
   <td><strong>As a registered user</strong>
   </td>
   <td><strong>I want my statistics to be recorded after each game</strong>
   </td>
   <td><strong>Upon receiving a </strong>POST request to /v1/users/{userID}, <strong>with a json object containing the number of correct answers, whether or not they won, and how many points they received for the game, and StatsHandler will insert that information into UserStatistics </strong>
   </td>
  </tr>
  <tr>
   <td><strong>P3</strong>
   </td>
   <td><strong>As a registered user</strong>
   </td>
   <td><strong>I want to view my statistics, or another player's statistics</strong>
   </td>
   <td><strong>Upon receiving a </strong>GET request to /v1/users/{userID}, <strong>StatsHandler will query the UserStatistics table for each with provided userID, and return a json object containing the sum for each statistic category</strong>
   </td>
  </tr>
  <tr>
   <td><strong>P2</strong>
   </td>
   <td><strong>As a registered user</strong>
   </td>
   <td><strong>I want to send chat messages to other players in my game</strong>
   </td>
   <td><strong>Upon receiving a </strong>POST request to /v1/trivia/{triviaID}/messages, <strong>the messaging microservice will insert the message body and the associated triviaID into the Message table</strong>
   </td>
  </tr>
  <tr>
   <td><strong>P2</strong>
   </td>
   <td><strong>As a registered user</strong>
   </td>
   <td><strong>I want to view chat messages sent by other players in my game</strong>
   </td>
   <td><strong>Upon receiving a </strong>GET request to /v1/trivia/{triviaID}/messages, <strong>the messaging microservice will respond with a list of all the messages for that game</strong>
   </td>
  </tr>
</table>



# Schemas 

Database 



<p id="gdcalert3" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Project-proposal2.jpg). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert4">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Project-proposal2.jpg "image_tooltip")



# Endpoints

The trivia microservice will handle the state of the game. The state will include a list of questions to ask and the users currently in the game. The service will generate a set of questions to ask users based on the API we are pulling from. After each question is answered by users in the alloted time those that didn't answer correctly will be removed from the game. 

POST /v1/users



*   Given a json object containing user information, including email, password, first name, last name, user name, and a photoURL, inserts the user into the user store
*   Returns the newly inserted user, along with their ID
*   Creates a session for the newly inserted user in the session store

GET /v1/users/{userID}



*   Given a userID, will return the user name and statistics for that specific userID 

POST /v1/user/{userID}



*   Adds user game statistics to UserStatistics table after a game has been completed

POST /v1/trivia



*   Insert a new trivia game into the trivia microservice

POST /v1/trivia/{triviaID}



*   Add user to lobby
    *   Upgrade to websocket to send questions
*   Start game if time runs out on lobby waiting

GET /v1/trivia/{triviaID}/answer



*   User sends answer to question
*   Game state updated
*   If answer incorrect user removed from state
*   Send user result of answer

POST /v1/trivia/{triviaID}/messages



*   Inserts the message body and the creator of the message into the messaging microservice

GET /v1/trivia/{triviaID}/message



*   Returns all messages for the request trivia game, along with their creators

We will also utilize the API endpoints we defined in assignments throughout the quarter for signing users up, storing session information for active users, and sending/receiving messages. 
