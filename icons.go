package main

type iconInfo struct {
	Emoji string
	Name  string
	ID    int64
}

var icons = []iconInfo{
	{ID: 0, Emoji: "🔑", Name: "Key"},
	{ID: 1, Emoji: "🌐", Name: "World"},
	{ID: 2, Emoji: "⚠️", Name: "Warning"},
	{ID: 3, Emoji: "📡", Name: "Network"},
	{ID: 4, Emoji: "📄", Name: "Document"},
	{ID: 5, Emoji: "📁", Name: "Folder"},
	{ID: 6, Emoji: "📂", Name: "Open Folder"},
	{ID: 7, Emoji: "💻", Name: "Computer"},
	{ID: 8, Emoji: "📋", Name: "Clipboard"},
	{ID: 9, Emoji: "🪪", Name: "Identity"},
	{ID: 10, Emoji: "🗄️", Name: "Database"},
	{ID: 11, Emoji: "🧰", Name: "Tools"},
	{ID: 12, Emoji: "📷", Name: "Camera"},
	{ID: 13, Emoji: "💾", Name: "Storage"},
	{ID: 14, Emoji: "🔌", Name: "Connection"},
	{ID: 15, Emoji: "🎙️", Name: "Audio"},
	{ID: 16, Emoji: "🎬", Name: "Video"},
	{ID: 17, Emoji: "📦", Name: "Package"},
	{ID: 18, Emoji: "📜", Name: "Certificate"},
	{ID: 19, Emoji: "📧", Name: "Email"},
	{ID: 20, Emoji: "👤", Name: "User"},
	{ID: 21, Emoji: "👥", Name: "Users"},
	{ID: 22, Emoji: "🔒", Name: "Locked"},
	{ID: 23, Emoji: "🔓", Name: "Unlocked"},
	{ID: 24, Emoji: "⚙️", Name: "Settings"},
	{ID: 25, Emoji: "📝", Name: "Notes"},
	{ID: 26, Emoji: "🔍", Name: "Search"},
	{ID: 27, Emoji: "⭐", Name: "Favorite"},
	{ID: 28, Emoji: "📌", Name: "Pinned"},
	{ID: 29, Emoji: "⏰", Name: "Time"},
	{ID: 30, Emoji: "📅", Name: "Calendar"},
	{ID: 31, Emoji: "🔔", Name: "Notification"},
	{ID: 32, Emoji: "❌", Name: "Error"},
	{ID: 33, Emoji: "✔️", Name: "Success"},
	{ID: 34, Emoji: "🔄", Name: "Sync"},
	{ID: 35, Emoji: "☁️", Name: "Cloud"},
	{ID: 36, Emoji: "📡", Name: "Network Signal"},
	{ID: 37, Emoji: "📱", Name: "Mobile"},
	{ID: 38, Emoji: "💳", Name: "Payment"},
	{ID: 39, Emoji: "🏦", Name: "Bank"},
	{ID: 40, Emoji: "💼", Name: "Work"},
	{ID: 41, Emoji: "🏠", Name: "Home"},
	{ID: 42, Emoji: "🎓", Name: "Study"},
	{ID: 43, Emoji: "🗑️", Name: "Trash"},
	{ID: 44, Emoji: "📱", Name: "Social Media"},
	{ID: 45, Emoji: "🧠", Name: "Knowledge"},
	{ID: 46, Emoji: "📊", Name: "Stats"},
	{ID: 47, Emoji: "🔐", Name: "Encryption"},
	{ID: 48, Emoji: "🧩", Name: "Plugin"},
	{ID: 49, Emoji: "🔥", Name: "Critical"},
	{ID: 50, Emoji: "🧊", Name: "Archive"},
	{ID: 51, Emoji: "🕹️", Name: "Gaming"},
}


func iconEmoji(id int64) string {
	if id >= 0 && id < int64(len(icons)) {
		return icons[id].Emoji
	}
	return "❓"
}

func iconName(id int64) string {
	if id >= 0 && id < int64(len(icons)) {
		return icons[id].Name
	}
	return "Unknown"
}
