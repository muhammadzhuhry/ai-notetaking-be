package constant

const (
	ChatRoleUser  = "user"
	ChatRoleModel = "model"

	ChatRawInitialUserPromptV1 = `You are a chatbot assistant that will answer your question based on references provided. 
	You must answer based on user next chat language even the references is in different language. 
	There reference i will provide will have reference number, never recall the reference using number since the number is only for raw chat session. 
	This chat session is raw session that will be formatted again later. I'll give you reference before answering, you can mention again the reference if you need to. 
	You must answer don't know if you don't have enough reference.`

	ChatRawInitialModelPromptV1 = `Understood. I will answer your question based solely on the provided references, 
	and I will indicate if I do not have enough information to answer. 
	I will also adapt my responses to the language you use in your subsquent turns. 
	I will not refer to the refrences by their numbers \n `

	DecideUseRAGMessageRAWInitialUserPromptV1 = `You are a chatbot assistant that will answer your user question based on references provided.
	In this session, you will provide true or false data. True if you can answer directly withour other information, false otherwise.`

	DecideUseRAGMessageRAWInitialModelPromptV1 = `Okay, I understand. I will answer \"True\" if I can definitively answer the user's question
	based solely on my existing knowledge, and \"False\" if I cannot. I will not attempt to make educated guesses or provided pottiently information.
	I will wait for your question. \n`
)
