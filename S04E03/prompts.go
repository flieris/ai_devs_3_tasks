package main

const (
	ResponseFormat = `
<response_format>
{
  "_thinking": "explanation of your interpretation and decision process",
  "answers": ["List of answers in string format"]
}
</response_format>
`
	CategorizeLinksPrompt = `
<objective>
Bazując na linkach w <documents>. Skategoryzuj jakie linki byłyby najlepsze do odpowiedzi na dane pytanie. Odpowiedź zgodnie z formatem JSON (bez bloków markdown)
</objective>
<rules>
- Odpowiadaj TYLKO samymi linkami
- Produkty lub projekty dla jakiegoś klienta sa trzymane w portfoliach
- Możesz podać kilka linków w formie listy 
- Zwróc odpowiedź jako JSON zgodnie z polem <response_format>
- sformatuj odpowiedź zgodnie z polem <response_format> bez bloków markdownowych
- tematy odnoszące się do samej firmy SoftoAI MOGĄ znajdować się także w aktualnościach lub blogu
- możesz podać linku które mogą zamwierać interesujące informacje - odpowiedź nie musi być precyzyjna.
- możesz się domyślać który link podać
</rules>
`
	QuestionPrompt = `
<objective>
Bazując na wiedzy z <documents>. Odpowiedź na pytanie podane przez użytkownika. Odpowiadaj zgodnie z formatem JSON (bez bloków markdown)
</objective>
<rules>
- odpowieadaj zwięźle i konkretnie
- Bazuj na wiedzy w polu <documents>
- Zwróc odpowiedź jako JSON zgodnie z polem <response_format>
- sformatuj odpowiedź zgodnie z polem <response_format> bez bloków markdownowych
</rules>
  `
)
