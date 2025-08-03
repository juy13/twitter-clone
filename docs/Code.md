# Code implementation

Let's talk about the code a little. 

Go was used as the programming language for this project. I think that the first prototype by Dorsey was in Ruby (?) and now it's working using scala, java, c++ and other interesting languages. But for MVP GO is the best solution.

## Why GO?

So, first of all, because I know it, but also, it's the simplest solution for web development rn. And it has lot's of web related benefits and libraries done.

### Code Architecture

Idiomatic Go doesn't use any specific architecture like Hexagonal Architecture or Domain Driven Design (DDD). Instead, it uses a simple structure with packages and interfaces to organize code. We are not developing Java project :smile

But I tried to separate as much as I could to build it flexible and easy to understand. There are separations between object description and object implementation. This violates a little the Go idea, because normally interfaces and there implementations in Go go in the dame file (idiomatically), but as we are doing the project that can work with different services (redis can be changed to kafka, postgres to mysql, etc.). That's why everything is separated.

