package users

func Authenticate(content string, username string, password string) (UserRecord, GroupRecord, error) {
	user, ok, err := FindActiveUser(content, username)
	if err != nil {
		return UserRecord{}, GroupRecord{}, err
	}
	if !ok {
		return UserRecord{}, GroupRecord{}, ErrUserNotFound
	}
	if user.Password != password {
		return UserRecord{}, GroupRecord{}, ErrBadPassword
	}

	group, ok, err := FindActiveGroup(content, user.Group)
	if err != nil {
		return UserRecord{}, GroupRecord{}, err
	}
	if !ok {
		return UserRecord{}, GroupRecord{}, ErrGroupNotFound
	}
	return user, group, nil
}
