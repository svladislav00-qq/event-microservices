package main

import "github.com/99designs/gqlgen/graphql"

type Server struct {
	// authClient     *auth.Client
	// eventClient    *event.Client
	// attendeeClient *attendee.Client
}

func NewGraphQLServer(authUrl, eventUrl, attendeeUrl string) (*Server, error) {
	// authClient, err := auth.NewClient(authUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// eventClient, err := event.NewClient(eventUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// attendeeClient, err := attendee.NewClient(attendeeUrl)
	// if err != nil {
	// 	return nil, err
	// }

	return &Server{
		// authClient,
		// eventClient,
		// attendeeClient,
	}, nil
}

// func (s *Server) Mutation() MutationResolver {
// 	return &mutationResolver{
// 		server: s,
// 	}
// }

// func (s *Server) Query() QueryResolver {
// 	return &queryResolver{
// 		server: s,
// 	}
// }

// func (s *Server) Account() AccountResolver {
// 	return &accountResolver{
// 		server: s,
// 	}
// }

// func (s *Server) Department() DepartmentResolver {
// 	return &departmentResolver{
// 		server: s,
// 	}
// }

// func (s *Server) Event() EventResolver {
// 	return &eventResolver{
// 		server: s,
// 	}
// }

// func (s *Server) Attendee() AttendeeResolver {
// 	return &attendeeResolver{
// 		server: s,
// 	}
// }

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}
