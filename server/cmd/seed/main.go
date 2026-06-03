package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/migrate"
	_ "github.com/lib/pq"
)

type SeedIcon struct {
	Name       string   `json:"name"`
	SvgContent string   `json:"svg_content"`
	IsPublic   bool     `json:"is_public"`
	Tags       []SeedTag `json:"tags"`
	Theme      string   `json:"theme"`
}

type SeedTag struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type BatchRequest struct {
	Icons []SeedIcon `json:"icons"`
}

var lineIcon = func(paths string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#3B82F6" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">%s</svg>`, paths)
}

var filledIcon = func(shapes string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">%s</svg>`, shapes)
}

func tag(usageName, styleName string) []SeedTag {
	return []SeedTag{
		{Name: usageName, Type: "usage"},
		{Name: styleName, Type: "style"},
	}
}

func icons() []SeedIcon {
	return []SeedIcon{
		// --- LINE style (usage ×3 weight for style="line") ---
		{Name: "home-line.svg", SvgContent: lineIcon(`<path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>`), IsPublic: true, Tags: tag("home", "line"), Theme: "UI"},
		{Name: "user-line.svg", SvgContent: lineIcon(`<path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/>`), IsPublic: true, Tags: tag("user", "line"), Theme: "UI"},
		{Name: "search-line.svg", SvgContent: lineIcon(`<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>`), IsPublic: true, Tags: tag("search", "line"), Theme: "UI"},
		{Name: "settings-line.svg", SvgContent: lineIcon(`<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/>`), IsPublic: true, Tags: tag("settings", "line"), Theme: "UI"},
		{Name: "mail-line.svg", SvgContent: lineIcon(`<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>`), IsPublic: true, Tags: tag("mail", "line"), Theme: "UI"},
		{Name: "lock-line.svg", SvgContent: lineIcon(`<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>`), IsPublic: true, Tags: tag("lock", "line"), Theme: "UI"},
		{Name: "cart-line.svg", SvgContent: lineIcon(`<circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6"/>`), IsPublic: true, Tags: tag("cart", "line"), Theme: "UI"},
		{Name: "heart-line.svg", SvgContent: lineIcon(`<path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"/>`), IsPublic: true, Tags: tag("heart", "line"), Theme: "UI"},
		{Name: "star-line.svg", SvgContent: lineIcon(`<polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>`), IsPublic: true, Tags: tag("star", "line"), Theme: "UI"},
		{Name: "share-line.svg", SvgContent: lineIcon(`<circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/>`), IsPublic: true, Tags: tag("share", "line"), Theme: "UI"},
		{Name: "download-line.svg", SvgContent: lineIcon(`<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>`), IsPublic: true, Tags: tag("download", "line"), Theme: "UI"},
		{Name: "upload-line.svg", SvgContent: lineIcon(`<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>`), IsPublic: true, Tags: tag("upload", "line"), Theme: "UI"},
		{Name: "edit-line.svg", SvgContent: lineIcon(`<path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>`), IsPublic: true, Tags: tag("edit", "line"), Theme: "UI"},
		{Name: "delete-line.svg", SvgContent: lineIcon(`<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/>`), IsPublic: true, Tags: tag("delete", "line"), Theme: "UI"},
		{Name: "add-line.svg", SvgContent: lineIcon(`<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>`), IsPublic: true, Tags: tag("add", "line"), Theme: "UI"},
		{Name: "close-line.svg", SvgContent: lineIcon(`<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>`), IsPublic: true, Tags: tag("close", "line"), Theme: "UI"},
		{Name: "check-line.svg", SvgContent: lineIcon(`<polyline points="20 6 9 17 4 12"/>`), IsPublic: true, Tags: tag("check", "line"), Theme: "UI"},
		{Name: "arrow-right-line.svg", SvgContent: lineIcon(`<line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>`), IsPublic: true, Tags: tag("arrow", "line"), Theme: "UI"},
		{Name: "arrow-left-line.svg", SvgContent: lineIcon(`<line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/>`), IsPublic: true, Tags: tag("arrow", "line"), Theme: "UI"},
		{Name: "arrow-up-line.svg", SvgContent: lineIcon(`<line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/>`), IsPublic: true, Tags: tag("arrow", "line"), Theme: "UI"},
		{Name: "arrow-down-line.svg", SvgContent: lineIcon(`<line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/>`), IsPublic: true, Tags: tag("arrow", "line"), Theme: "UI"},
		{Name: "menu-line.svg", SvgContent: lineIcon(`<line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/>`), IsPublic: true, Tags: tag("menu", "line"), Theme: "UI"},
		{Name: "filter-line.svg", SvgContent: lineIcon(`<polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>`), IsPublic: true, Tags: tag("filter", "line"), Theme: "UI"},
		{Name: "bell-line.svg", SvgContent: lineIcon(`<path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/>`), IsPublic: true, Tags: tag("bell", "line"), Theme: "UI"},
		{Name: "bookmark-line.svg", SvgContent: lineIcon(`<path d="M19 21l-7-5-7 5V5a2 2 0 012-2h10a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("bookmark", "line"), Theme: "UI"},
		{Name: "camera-line.svg", SvgContent: lineIcon(`<path d="M23 19a2 2 0 01-2 2H3a2 2 0 01-2-2V8a2 2 0 012-2h4l2-3h6l2 3h4a2 2 0 012 2z"/><circle cx="12" cy="13" r="4"/>`), IsPublic: true, Tags: tag("camera", "line"), Theme: "UI"},
		{Name: "clock-line.svg", SvgContent: lineIcon(`<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>`), IsPublic: true, Tags: tag("clock", "line"), Theme: "UI"},
		{Name: "calendar-line.svg", SvgContent: lineIcon(`<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>`), IsPublic: true, Tags: tag("calendar", "line"), Theme: "UI"},
		{Name: "folder-line.svg", SvgContent: lineIcon(`<path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("folder", "line"), Theme: "UI"},
		{Name: "file-line.svg", SvgContent: lineIcon(`<path d="M13 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V9z"/><polyline points="13 2 13 9 20 9"/>`), IsPublic: true, Tags: tag("file", "line"), Theme: "UI"},
		{Name: "map-line.svg", SvgContent: lineIcon(`<polygon points="1 6 1 22 8 18 16 22 23 18 23 2 16 6 8 2 1 6"/><line x1="8" y1="2" x2="8" y2="18"/><line x1="16" y1="6" x2="16" y2="22"/>`), IsPublic: true, Tags: tag("map", "line"), Theme: "UI"},
		{Name: "phone-line.svg", SvgContent: lineIcon(`<path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07 19.5 19.5 0 01-6-6 19.79 19.79 0 01-3.07-8.67A2 2 0 014.11 2h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 16.92z"/>`), IsPublic: true, Tags: tag("phone", "line"), Theme: "UI"},
		{Name: "message-line.svg", SvgContent: lineIcon(`<path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("message", "line"), Theme: "UI"},
		{Name: "eye-line.svg", SvgContent: lineIcon(`<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>`), IsPublic: true, Tags: tag("eye", "line"), Theme: "UI"},
		{Name: "sun-line.svg", SvgContent: lineIcon(`<circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>`), IsPublic: true, Tags: tag("sun", "line"), Theme: "UI"},
		{Name: "moon-line.svg", SvgContent: lineIcon(`<path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>`), IsPublic: true, Tags: tag("moon", "line"), Theme: "UI"},
		{Name: "tag-line.svg", SvgContent: lineIcon(`<path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>`), IsPublic: true, Tags: tag("tag", "line"), Theme: "UI"},
		{Name: "link-line.svg", SvgContent: lineIcon(`<path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/>`), IsPublic: true, Tags: tag("link", "line"), Theme: "UI"},
		{Name: "video-line.svg", SvgContent: lineIcon(`<polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>`), IsPublic: true, Tags: tag("video", "line"), Theme: "UI"},
		{Name: "music-line.svg", SvgContent: lineIcon(`<path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>`), IsPublic: true, Tags: tag("music", "line"), Theme: "UI"},
		{Name: "compass-line.svg", SvgContent: lineIcon(`<circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/>`), IsPublic: true, Tags: tag("compass", "line"), Theme: "UI"},
		{Name: "cloud-line.svg", SvgContent: lineIcon(`<path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/>`), IsPublic: true, Tags: tag("cloud", "line"), Theme: "UI"},
		{Name: "flag-line.svg", SvgContent: lineIcon(`<path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" y1="22" x2="4" y2="15"/>`), IsPublic: true, Tags: tag("flag", "line"), Theme: "UI"},
		{Name: "globe-line.svg", SvgContent: lineIcon(`<circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>`), IsPublic: true, Tags: tag("globe", "line"), Theme: "UI"},
		{Name: "printer-line.svg", SvgContent: lineIcon(`<polyline points="6 9 6 2 18 2 18 9"/><path d="M6 12H4a2 2 0 00-2 2v4a2 2 0 002 2h16a2 2 0 002-2v-4a2 2 0 00-2-2h-2"/><rect x="6" y="14" width="12" height="8"/>`), IsPublic: true, Tags: tag("printer", "line"), Theme: "UI"},
		{Name: "database-line.svg", SvgContent: lineIcon(`<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>`), IsPublic: true, Tags: tag("database", "line"), Theme: "UI"},
		{Name: "monitor-line.svg", SvgContent: lineIcon(`<rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>`), IsPublic: true, Tags: tag("monitor", "line"), Theme: "UI"},

		// --- FILLED style ---
		{Name: "home-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline fill="none" stroke="#3B82F6" stroke-width="2" points="9 22 9 12 15 12 15 22"/>`), IsPublic: true, Tags: tag("home", "filled"), Theme: "UI"},
		{Name: "user-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle fill="#3B82F6" cx="12" cy="7" r="4"/>`), IsPublic: true, Tags: tag("user", "filled"), Theme: "UI"},
		{Name: "heart-filled.svg", SvgContent: filledIcon(`<path fill="#EF4444" d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"/>`), IsPublic: true, Tags: tag("heart", "filled"), Theme: "UI"},
		{Name: "star-filled.svg", SvgContent: filledIcon(`<polygon fill="#F59E0B" points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>`), IsPublic: true, Tags: tag("star", "filled"), Theme: "UI"},
		{Name: "lock-filled.svg", SvgContent: filledIcon(`<rect fill="#10B981" x="3" y="11" width="18" height="11" rx="2" ry="2"/><path fill="none" stroke="#10B981" stroke-width="2" d="M7 11V7a5 5 0 0110 0v4"/>`), IsPublic: true, Tags: tag("lock", "filled"), Theme: "UI"},
		{Name: "cart-filled.svg", SvgContent: filledIcon(`<circle fill="#6B7280" cx="9" cy="21" r="1"/><circle fill="#6B7280" cx="20" cy="21" r="1"/><path fill="#6B7280" d="M1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6"/>`), IsPublic: true, Tags: tag("cart", "filled"), Theme: "UI"},
		{Name: "bell-filled.svg", SvgContent: filledIcon(`<path fill="#F59E0B" d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path fill="#F59E0B" d="M13.73 21a2 2 0 01-3.46 0"/>`), IsPublic: true, Tags: tag("bell", "filled"), Theme: "UI"},
		{Name: "bookmark-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M19 21l-7-5-7 5V5a2 2 0 012-2h10a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("bookmark", "filled"), Theme: "UI"},
		{Name: "check-filled.svg", SvgContent: filledIcon(`<circle fill="#10B981" cx="12" cy="12" r="10"/><polyline fill="none" stroke="#fff" stroke-width="2" points="8 12 11 15 16 9"/>`), IsPublic: true, Tags: tag("check", "filled"), Theme: "UI"},
		{Name: "close-filled.svg", SvgContent: filledIcon(`<circle fill="#EF4444" cx="12" cy="12" r="10"/><line stroke="#fff" stroke-width="2" x1="8" y1="8" x2="16" y2="16"/><line stroke="#fff" stroke-width="2" x1="16" y1="8" x2="8" y2="16"/>`), IsPublic: true, Tags: tag("close", "filled"), Theme: "UI"},
		{Name: "add-filled.svg", SvgContent: filledIcon(`<circle fill="#3B82F6" cx="12" cy="12" r="10"/><line stroke="#fff" stroke-width="2" x1="12" y1="7" x2="12" y2="17"/><line stroke="#fff" stroke-width="2" x1="7" y1="12" x2="17" y2="12"/>`), IsPublic: true, Tags: tag("add", "filled"), Theme: "UI"},
		{Name: "sun-filled.svg", SvgContent: filledIcon(`<circle fill="#F59E0B" cx="12" cy="12" r="5"/><g stroke="#F59E0B" stroke-width="2" stroke-linecap="round"><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></g>`), IsPublic: true, Tags: tag("sun", "filled"), Theme: "UI"},
		{Name: "folder-filled.svg", SvgContent: filledIcon(`<path fill="#F59E0B" d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("folder", "filled"), Theme: "UI"},
		{Name: "camera-filled.svg", SvgContent: filledIcon(`<path fill="#6B7280" d="M23 19a2 2 0 01-2 2H3a2 2 0 01-2-2V8a2 2 0 012-2h4l2-3h6l2 3h4a2 2 0 012 2z"/><circle fill="#6B7280" cx="12" cy="13" r="4"/>`), IsPublic: true, Tags: tag("camera", "filled"), Theme: "UI"},

		// --- Additional FILLED style (batch 2) ---
		{Name: "search-filled.svg", SvgContent: filledIcon(`<circle fill="#3B82F6" cx="11" cy="11" r="8"/><line stroke="#fff" stroke-width="2" x1="21" y1="21" x2="16.65" y2="16.65"/>`), IsPublic: true, Tags: tag("search", "filled"), Theme: "UI"},
		{Name: "settings-filled.svg", SvgContent: filledIcon(`<circle fill="#6B7280" cx="12" cy="12" r="3"/><path fill="#6B7280" d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/>`), IsPublic: true, Tags: tag("settings", "filled"), Theme: "UI"},
		{Name: "mail-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline fill="none" stroke="#fff" stroke-width="2" points="22,6 12,13 2,6"/>`), IsPublic: true, Tags: tag("mail", "filled"), Theme: "UI"},
		{Name: "download-filled.svg", SvgContent: filledIcon(`<circle fill="#10B981" cx="12" cy="12" r="10"/><polyline fill="none" stroke="#fff" stroke-width="2" points="8 12 12 16 16 12"/><line stroke="#fff" stroke-width="2" x1="12" y1="8" x2="12" y2="16"/>`), IsPublic: true, Tags: tag("download", "filled"), Theme: "UI"},
		{Name: "upload-filled.svg", SvgContent: filledIcon(`<circle fill="#8B5CF6" cx="12" cy="12" r="10"/><polyline fill="none" stroke="#fff" stroke-width="2" points="16 8 12 4 8 8"/><line stroke="#fff" stroke-width="2" x1="12" y1="4" x2="12" y2="16"/>`), IsPublic: true, Tags: tag("upload", "filled"), Theme: "UI"},
		{Name: "edit-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path fill="#3B82F6" d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>`), IsPublic: true, Tags: tag("edit", "filled"), Theme: "UI"},
		{Name: "cloud-filled.svg", SvgContent: filledIcon(`<path fill="#3B82F6" d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/>`), IsPublic: true, Tags: tag("cloud", "filled"), Theme: "UI"},
		{Name: "globe-filled.svg", SvgContent: filledIcon(`<circle fill="#3B82F6" cx="12" cy="12" r="10"/><line stroke="#fff" stroke-width="2" x1="2" y1="12" x2="22" y2="12"/><path fill="#3B82F6" d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>`), IsPublic: true, Tags: tag("globe", "filled"), Theme: "UI"},
		{Name: "flag-filled.svg", SvgContent: filledIcon(`<path fill="#EF4444" d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line stroke="#EF4444" stroke-width="2" x1="4" y1="22" x2="4" y2="15"/>`), IsPublic: true, Tags: tag("flag", "filled"), Theme: "UI"},
		{Name: "music-filled.svg", SvgContent: filledIcon(`<path fill="#8B5CF6" d="M9 18V5l12-2v13"/><circle fill="#8B5CF6" cx="6" cy="18" r="3"/><circle fill="#8B5CF6" cx="18" cy="16" r="3"/>`), IsPublic: true, Tags: tag("music", "filled"), Theme: "UI"},
		{Name: "tag-filled.svg", SvgContent: filledIcon(`<path fill="#F59E0B" d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z"/><circle fill="#F59E0B" cx="7" cy="7" r="1"/>`), IsPublic: true, Tags: tag("tag", "filled"), Theme: "UI"},
		{Name: "compass-filled.svg", SvgContent: filledIcon(`<circle fill="#6B7280" cx="12" cy="12" r="10"/><polygon fill="#fff" points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/>`), IsPublic: true, Tags: tag("compass", "filled"), Theme: "UI"},
		{Name: "clock-filled.svg", SvgContent: filledIcon(`<circle fill="#6B7280" cx="12" cy="12" r="10"/><polyline stroke="#fff" stroke-width="2" fill="none" points="12 6 12 12 16 14"/>`), IsPublic: true, Tags: tag("clock", "filled"), Theme: "UI"},
		{Name: "calendar-filled.svg", SvgContent: filledIcon(`<rect fill="#3B82F6" x="3" y="4" width="18" height="18" rx="2" ry="2"/><line stroke="#fff" stroke-width="2" x1="16" y1="2" x2="16" y2="6"/><line stroke="#fff" stroke-width="2" x1="8" y1="2" x2="8" y2="6"/><line stroke="#fff" stroke-width="2" x1="3" y1="10" x2="21" y2="10"/>`), IsPublic: true, Tags: tag("calendar", "filled"), Theme: "UI"},
		{Name: "message-filled.svg", SvgContent: filledIcon(`<path fill="#10B981" d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>`), IsPublic: true, Tags: tag("message", "filled"), Theme: "UI"},
		{Name: "eye-filled.svg", SvgContent: filledIcon(`<path fill="#6B7280" d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle fill="#6B7280" cx="12" cy="12" r="3"/>`), IsPublic: true, Tags: tag("eye", "filled"), Theme: "UI"},
		{Name: "share-filled.svg", SvgContent: filledIcon(`<circle fill="#3B82F6" cx="18" cy="5" r="3"/><circle fill="#3B82F6" cx="6" cy="12" r="3"/><circle fill="#3B82F6" cx="18" cy="19" r="3"/><line stroke="#3B82F6" stroke-width="2" x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line stroke="#3B82F6" stroke-width="2" x1="15.41" y1="6.51" x2="8.59" y2="10.49"/>`), IsPublic: true, Tags: tag("share", "filled"), Theme: "UI"},
		{Name: "map-filled.svg", SvgContent: filledIcon(`<polygon fill="#10B981" points="1 6 1 22 8 18 16 22 23 18 23 2 16 6 8 2 1 6"/><line stroke="#fff" stroke-width="1.5" x1="8" y1="2" x2="8" y2="18"/><line stroke="#fff" stroke-width="1.5" x1="16" y1="6" x2="16" y2="22"/>`), IsPublic: true, Tags: tag("map", "filled"), Theme: "UI"},
		{Name: "phone-filled.svg", SvgContent: filledIcon(`<path fill="#10B981" d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07 19.5 19.5 0 01-6-6 19.79 19.79 0 01-3.07-8.67A2 2 0 014.11 2h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 16.92z"/>`), IsPublic: true, Tags: tag("phone", "filled"), Theme: "UI"},
		{Name: "video-filled.svg", SvgContent: filledIcon(`<polygon fill="#EF4444" points="23 7 16 12 23 17 23 7"/><rect fill="#EF4444" x="1" y="5" width="15" height="14" rx="2" ry="2"/>`), IsPublic: true, Tags: tag("video", "filled"), Theme: "UI"},
		{Name: "file-filled.svg", SvgContent: filledIcon(`<path fill="#6B7280" d="M13 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V9z"/><polyline fill="none" stroke="#fff" stroke-width="2" points="13 2 13 9 20 9"/>`), IsPublic: true, Tags: tag("file", "filled"), Theme: "UI"},
	}
}

func tags() []SeedTag {
	return []SeedTag{
		{Name: "home", Type: "usage"}, {Name: "user", Type: "usage"}, {Name: "search", Type: "usage"},
		{Name: "settings", Type: "usage"}, {Name: "mail", Type: "usage"}, {Name: "lock", Type: "usage"},
		{Name: "cart", Type: "usage"}, {Name: "heart", Type: "usage"}, {Name: "star", Type: "usage"},
		{Name: "share", Type: "usage"}, {Name: "download", Type: "usage"}, {Name: "upload", Type: "usage"},
		{Name: "edit", Type: "usage"}, {Name: "delete", Type: "usage"}, {Name: "add", Type: "usage"},
		{Name: "close", Type: "usage"}, {Name: "check", Type: "usage"}, {Name: "arrow", Type: "usage"},
		{Name: "menu", Type: "usage"}, {Name: "filter", Type: "usage"}, {Name: "bell", Type: "usage"},
		{Name: "bookmark", Type: "usage"}, {Name: "camera", Type: "usage"}, {Name: "clock", Type: "usage"},
		{Name: "calendar", Type: "usage"}, {Name: "folder", Type: "usage"}, {Name: "file", Type: "usage"},
		{Name: "map", Type: "usage"}, {Name: "phone", Type: "usage"}, {Name: "message", Type: "usage"},
		{Name: "eye", Type: "usage"}, {Name: "sun", Type: "usage"}, {Name: "moon", Type: "usage"},
		{Name: "tag", Type: "usage"}, {Name: "link", Type: "usage"}, {Name: "video", Type: "usage"},
		{Name: "music", Type: "usage"}, {Name: "compass", Type: "usage"}, {Name: "cloud", Type: "usage"},
		{Name: "flag", Type: "usage"}, {Name: "globe", Type: "usage"}, {Name: "printer", Type: "usage"},
		{Name: "database", Type: "usage"}, {Name: "monitor", Type: "usage"},
		{Name: "line", Type: "style"}, {Name: "filled", Type: "style"},
		{Name: "UI", Type: "category"},
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: seed <api_base_url> <auth_token>")
		fmt.Println("Example: seed http://localhost:8080 <guest-token>")
		os.Exit(1)
	}

	baseURL := os.Args[1]
	token := os.Args[2]

	allIcons := icons()

	// Split into batches of 50
	for i := 0; i < len(allIcons); i += 50 {
		end := i + 50
		if end > len(allIcons) {
			end = len(allIcons)
		}
		batch := allIcons[i:end]
		body, _ := json.Marshal(BatchRequest{Icons: batch})
		req, _ := http.NewRequest("POST", baseURL+"/api/v1/icons/batch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("batch %d: request failed: %v", i/50+1, err)
		}
		resp.Body.Close()
		if resp.StatusCode != 201 {
			log.Fatalf("batch %d: got status %d", i/50+1, resp.StatusCode)
		}
		log.Printf("batch %d: %d icons seeded", i/50+1, len(batch))
	}
	log.Printf("Done! %d icons seeded", len(allIcons))

	// Also migrate if DATABASE_URL is set
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err == nil {
			if err := db.Ping(); err == nil {
				_ = migrate.Run(db, "migrations")
				log.Println("migrations applied")
			}
			db.Close()
		}
	}
}
