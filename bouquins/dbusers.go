package bouquins

// Account returns user account from authentifier
func Account(authentifier string) (*UserAccount, error) {
	account := new(UserAccount)
	err := stmtAccount.QueryRow(authentifier).Scan(&account.ID, &account.DisplayName)
	if err != nil {
		return nil, err
	}
	return account, nil
}
