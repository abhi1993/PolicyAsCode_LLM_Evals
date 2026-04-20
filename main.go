package main                                                                                                                                                                                
                                                                                                                                                                                              
  import (
      "context"                                                                                                                                                                               
      "fmt"                                                                                                                                                                                 
      "log"

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
         option.WithAPIKey(""), // or set ANTHROPIC_API_KEY env var
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
                                                                                                                                                                                            
      // Tool 3: Evaluate Policy
      s.AddTool(mcp.NewTool("evaluate_policy",
          mcp.WithDescription("Evaluate whether a resource would pass or fail a given policy"),                                                                                               
          mcp.WithString("policy", mcp.Required(), mcp.Description("The policy to evaluate")),                                                                                                
          mcp.WithString("resource", mcp.Required(), mcp.Description("The resource or request to evaluate against")),                                                                         
      ), evaluatePolicyHandler)                                                                                                                                                               
                                                                                                                                                                                            
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

func evaluatePolicyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {                                                                                     
   args := req.Params.Arguments.(map[string]any)
   _ = args["policy"].(string)
   _ = args["resource"].(string)
   return mcp.NewToolResultText("Evaluation result: PASS (stub)"), nil                                                                                                                     
}    