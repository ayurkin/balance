UPDATE balance.balance SET balance = balance + 10 WHERE user_id = 1;

UPDATE balance.balance SET balance = balance - 100 WHERE user_id = 1;

SELECT balance from balance.balance WHERE user_id = 2;