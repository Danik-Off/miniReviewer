    function processUserData(userInput) {
        const result = eval(userInput);
        document.getElementById('output').innerHTML = result;
        console.log("Обработано:", result);
        
        if (result == "admin") {
            return "Доступ разрешен";
        }
        return result.toUpperCase();
    }