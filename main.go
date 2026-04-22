package main                                                                                                                                                                                
                                                                                                                                                                                              
  import (
      "context"                                                                                                                                                                               
      "fmt"                                                                                                                                                                                 
      "log"
      "os"

      "github.com/anthropics/anthropic-sdk-go"
      "github.com/anthropics/anthropic-sdk-go/option"
      "github.com/mark3labs/mcp-go/mcp"
      "github.com/mark3labs/mcp-go/server"
  )

  var anthropicClient anthropic.Client
   
  func main() {                                                                                                                                                                               
      s := server.NewMCPServer(                                                                                                                                                             
          "cedar-kyverno-policy-server",
          "1.0.0",
          server.WithToolCapabilities(false),
      )

      anthropicClient = anthropic.NewClient(
        option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
     )
      // Tool 1: Generate Policy                                                                                                                                                              
      s.AddTool(mcp.NewTool("generate_policy",                                                                                                                                              
          mcp.WithDescription("Generate a Kyverno or Cedar policy from a natural language description"),                                                                                      
          mcp.WithString("description", mcp.Required(), mcp.Description("What the policy should do")),
          mcp.WithString("type", mcp.Description("Policy type: 'kyverno' or 'cedar'")),                                                                                                       
      ), generatePolicyHandler)                                                                                                                                                               
                                                                                                                                                                                              
      // Tool 2: Explain Policy                                                                                                                                                               
      s.AddTool(mcp.NewTool("explain_policy",                                                                                                                                               
          mcp.WithDescription("Explain what a given policy does in plain English"),                                                                                                           
          mcp.WithString("policyContents", mcp.Required(), mcp.Description("The policy YAML or text to explain")),
      ), explainPolicyHandler)

      // Tool 3: Generate policy tests
      s.AddTool(mcp.NewTool("generate_policy_tests",
          mcp.WithDescription("Generate policy tests"),
          mcp.WithString("prompt", mcp.Required(), mcp.Description("English language description of the policy to evaluate")),
          ), generatePolicyTestsHandler)

      // Tool 4: Run policy tests
      s.AddTool(mcp.NewTool("run_kyverno_est",
          mcp.WithDescription("Run a kyverno policy tests"),
          mcp.WithString("prompt", mcp.Required(), mcp.Description("Given a policy, policy tests and sample resource file run the test.")),
          ), runPolicyTestsHandler)

      if err := server.ServeStdio(s); err != nil {
          log.Fatal(err)
      }                                                                                                                                                                                       
  }                                                                                                                                                                                         

  func generatePolicyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {                                                                                     
   args := req.Params.Arguments.(map[string]any)                                                                                                                                           
   description := args["description"].(string)                                                                                                                                             
   policyType, _ := args["type"].(string)                                                                                                                                                  
   if policyType == "" {
       policyType = "kyverno"                                                                                                                                                              
   }

   msg, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
      Model:     anthropic.ModelClaudeSonnet4_5,
      MaxTokens: 2048,
      System: []anthropic.TextBlockParam{
         {Text: "You are a Kubernetes policy expert. Return only valid YAML, no explanation."},
      },
      Messages: []anthropic.MessageParam{
         anthropic.NewUserMessage(anthropic.NewTextBlock(
            fmt.Sprintf("Generate a %s policy that does the following: %s", policyType, description),
            )),
         },
   })

   if err != nil {
      return nil, err
   }
   
   result := fmt.Sprintf("Generated %s policy for: %s\n(stub)", policyType, msg)
   return mcp.NewToolResultText(result), nil
}               
                                                                                                                                                                                           
func explainPolicyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {                                                                                      
   args := req.Params.Arguments.(map[string]any)
   policyContents := args["policyContents"].(string)
   msg, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
      Model:     anthropic.ModelClaudeSonnet4_5,
      MaxTokens: 2048,
      System: []anthropic.TextBlockParam{
         {Text: "You are a Kubernetes policy expert. Please explain the meaning of each policy sent to you."},
      },
      Messages: []anthropic.MessageParam{
         anthropic.NewUserMessage(anthropic.NewTextBlock(
            fmt.Sprintf("Please explain what the following policy does: %s", policyContents),
            )),
         },
   })

   if err != nil {
      return nil, err
   }

   result := fmt.Sprintf("The policy does the following:%s\n", msg)
   return mcp.NewToolResultText(result), nil
   }

func generatePolicyTestsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
   args := req.Params.Arguments.(map[string]any)
   englishPrompt := args["prompt"].(string)
   msg, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
      Model:     anthropic.ModelClaudeSonnet4_5,
      MaxTokens: 2048,
      System: []anthropic.TextBlockParam{
         {Text: "You are a kyverno expert. When given an english language prompt description for a policy, generate a suite of 10 tests for it. Return a kyverno-test.yaml + JSON list of resource files."},
      },
      Messages: []anthropic.MessageParam{
         anthropic.NewUserMessage(anthropic.NewTextBlock(
            fmt.Sprintf("Please generate a test suite for a policy that does the following: %s. In your suite also include sample resources.),", englishPrompt),
            )),
         },
   })

   if err != nil {
      return nil, err
   }

   result := fmt.Sprintf("JSON list of tests and resources:%s\n", msg)
   return mcp.NewToolResultText(result), nil
}

func evaluatePolicyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {                                                                                     
   args := req.Params.Arguments.(map[string]any)
   englishPrompt := args["prompt"].(string)
   msg, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
      Model:     anthropic.ModelClaudeSonnet4_5,
      MaxTokens: 2048,
      System: []anthropic.TextBlockParam{
         {Text: "You are a kyverno expert. When given an english language prompt description for a policy, generate a suite of 10 tests for it."},
      },
      Messages: []anthropic.MessageParam{
         anthropic.NewUserMessage(anthropic.NewTextBlock(
            fmt.Sprintf("Please generate a test suite for a policy that does the following: %s. In your suite also include sample resources. Return it as a JSON list with key value pairs of, test and resource yamls.", englishPrompt),
            )),
         },
   })

   if err != nil {
      return nil, err
   }

   result := fmt.Sprintf("JSON list of tests and resources:%s\n", msg)
   return mcp.NewToolResultText(result), nil
}

func runPolicyTestsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
   args := req.Params.Arguments.(map[string]any)
   englishPrompt := args["prompt"].(string)
   msg, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
      Model:     anthropic.ModelClaudeSonnet4_5,
      MaxTokens: 2048,
      System: []anthropic.TextBlockParam{
         {Text: "You are a kyverno expert. When given an english language prompt description for a policy, generate a suite of 10 tests for it."},
      },
      Messages: []anthropic.MessageParam{
         anthropic.NewUserMessage(anthropic.NewTextBlock(
            fmt.Sprintf("Please generate a test suite for a policy that does the following: %s. In your suite also include sample resources. Return it as a JSON list with key value pairs of, test and resource yamls.", englishPrompt),
            )),
         },
   })

   if err != nil {
      return nil, err
   }

   result := fmt.Sprintf("JSON list of tests and resources:%s\n", msg)
   return mcp.NewToolResultText(result), nil
}