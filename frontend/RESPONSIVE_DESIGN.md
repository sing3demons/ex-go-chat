# Responsive Design

This document describes the responsive design implementation for the chat application, ensuring optimal user experience across mobile, tablet, and desktop devices.

## Overview

The application uses a mobile-first responsive design approach with Tailwind CSS breakpoints:
- **Mobile**: < 640px (sm)
- **Tablet**: 640px - 1024px (sm to lg)
- **Desktop**: ≥ 1024px (lg+)

## Breakpoints

Tailwind CSS breakpoints used throughout the application:
- `sm`: 640px - Small devices (large phones)
- `md`: 768px - Medium devices (tablets)
- `lg`: 1024px - Large devices (desktops)
- `xl`: 1280px - Extra large devices (large desktops)

## Layout Adaptations

### Mobile Layout (< 1024px)

**Features:**
- Hamburger menu for room list
- Full-screen chat view when room selected
- Overlay sidebar that slides in from left
- Touch-optimized buttons (larger tap targets)
- Compact header with essential actions only
- Hidden non-essential UI elements

**Behavior:**
- Room list hidden by default
- Tap hamburger menu to show room list
- Tap room to select and auto-close sidebar
- Tap overlay to close sidebar
- Chat takes full screen width

**Components:**
- `ChatPage`: Mobile menu toggle, overlay, fixed sidebar
- `RoomList`: Full-width sidebar on mobile
- `ChatWindow`: Full-screen chat area
- `MessageInput`: Compact buttons, hidden emoji picker
- `MessageList`: Reduced padding

### Tablet Layout (640px - 1024px)

**Features:**
- Similar to mobile but with more breathing room
- Slightly larger text and spacing
- More UI elements visible
- Better use of available space

**Behavior:**
- Still uses hamburger menu for room list
- Wider sidebar when open
- More comfortable touch targets
- Connection status visible on medium+ screens

**Components:**
- `RoomList`: Width adjusts (w-80 on lg, w-96 on xl)
- `ChatWindow`: Better spacing
- Header: More elements visible

### Desktop Layout (≥ 1024px)

**Features:**
- Persistent sidebar (always visible)
- Multi-column layout
- Hover states for better interactivity
- All UI elements visible
- Optimal use of screen real estate

**Behavior:**
- Room list always visible on left
- Chat area takes remaining space
- No hamburger menu
- All actions and status indicators visible

**Components:**
- `RoomList`: Fixed width sidebar (320px - 384px)
- `ChatWindow`: Flexible width
- Header: Full feature set visible

## Component-Specific Responsive Features

### ChatPage

**Mobile:**
```tsx
- Hamburger menu button (lg:hidden)
- Fixed sidebar with overlay
- Compact header spacing (px-2 sm:px-4)
- Smaller icons (w-5 h-5 sm:w-6 sm:h-6)
```

**Desktop:**
```tsx
- No hamburger menu (hidden lg:block)
- Relative sidebar positioning
- Full header spacing
- Standard icon sizes
```

### RoomList

**Responsive Width:**
```tsx
className="w-full lg:w-80 xl:w-96"
```

**Features:**
- Full width on mobile
- 320px on desktop (lg)
- 384px on extra large screens (xl)
- Responsive padding (p-3 sm:p-4)
- Adaptive icon sizes

### RoomItem

**Touch Optimization:**
```tsx
className="p-3 sm:p-4 touch-manipulation active:bg-gray-300"
```

**Features:**
- Larger padding on mobile for better touch targets
- Active state for touch feedback
- Truncated text to prevent overflow
- Responsive font sizes (text-sm sm:text-base)

### ChatWindow

**Header Responsiveness:**
```tsx
- Compact padding (p-3 sm:p-4)
- Truncated room names
- Responsive font sizes (text-base sm:text-lg)
- Hidden actions on small screens
```

### MessageList

**Adaptive Spacing:**
```tsx
className="p-2 sm:p-4"
```

**Features:**
- Reduced padding on mobile
- Responsive empty state icons
- Adaptive date separator spacing
- Smaller text on mobile

### MessageInput

**Mobile Optimizations:**
```tsx
- Hidden emoji button on mobile (hidden sm:block)
- Compact button padding (px-3 sm:px-6)
- Smaller icons (w-4 h-4 sm:w-5 sm:h-5)
- Hidden hint text on mobile
- Touch-optimized buttons (touch-manipulation)
```

**Features:**
- Icon-only send button on mobile
- Text label on desktop
- Responsive textarea padding
- Adaptive button sizes

### Header

**Responsive Elements:**
```tsx
- Logo: w-7 h-7 sm:w-8 sm:h-8
- Title: hidden sm:block
- Connection status: hidden md:flex
- Username: hidden lg:block
- Room settings: hidden md:block
```

**Progressive Enhancement:**
- Essential actions always visible
- Secondary actions show on larger screens
- Status indicators appear when space allows

## Touch Optimization

### Touch Targets

All interactive elements meet minimum touch target size (44x44px):
```tsx
className="p-2"  // Minimum 44px height with icon
className="touch-manipulation"  // Prevents double-tap zoom
className="active:bg-gray-200"  // Visual feedback on touch
```

### Touch-Friendly Features

1. **Larger Tap Areas:**
   - Buttons have adequate padding
   - List items are tall enough
   - Icons are appropriately sized

2. **Touch Feedback:**
   - Active states on press
   - Hover states for desktop
   - Smooth transitions

3. **Gesture Support:**
   - Swipe to close sidebar (via overlay tap)
   - Scroll momentum in message list
   - Pull-to-refresh ready (not implemented)

## Typography Scaling

### Font Sizes

**Mobile:**
- Headers: text-base to text-lg
- Body: text-sm to text-base
- Small text: text-xs

**Desktop:**
- Headers: text-lg to text-xl
- Body: text-base
- Small text: text-xs to text-sm

### Responsive Text Classes

```tsx
text-xs sm:text-sm      // Small text
text-sm sm:text-base    // Body text
text-base sm:text-lg    // Subheadings
text-lg sm:text-xl      // Headings
```

## Spacing System

### Padding/Margin Scale

```tsx
p-2 sm:p-4              // Component padding
gap-1 sm:gap-2          // Flex gaps
mb-2 sm:mb-4            // Margins
```

### Consistent Spacing

- Mobile: Tighter spacing (2, 3)
- Tablet: Medium spacing (3, 4)
- Desktop: Comfortable spacing (4, 6)

## Visibility Classes

### Show/Hide Elements

```tsx
hidden sm:block         // Show on small+
hidden md:flex          // Show on medium+
hidden lg:block         // Show on large+
lg:hidden               // Hide on large+
```

### Common Patterns

**Mobile-only:**
```tsx
className="lg:hidden"
```

**Desktop-only:**
```tsx
className="hidden lg:block"
```

**Tablet+:**
```tsx
className="hidden sm:block"
```

## Layout Patterns

### Sidebar Pattern

```tsx
<aside className={`
  ${isMobileMenuOpen ? 'block' : 'hidden'} lg:block
  fixed lg:relative
  inset-0 lg:inset-auto
  z-20 lg:z-0
`}>
```

**Behavior:**
- Fixed overlay on mobile
- Relative positioning on desktop
- Conditional visibility
- Z-index management

### Flex Layouts

```tsx
className="flex gap-1 sm:gap-2 items-center"
```

**Features:**
- Responsive gaps
- Proper alignment
- Flex wrapping when needed
- Min-width constraints

## Best Practices

### 1. Mobile-First Approach

Start with mobile styles, then add larger breakpoints:
```tsx
// ✅ Good
className="text-sm sm:text-base lg:text-lg"

// ❌ Bad
className="lg:text-lg md:text-base text-sm"
```

### 2. Touch Targets

Ensure minimum 44x44px touch targets:
```tsx
// ✅ Good
className="p-2 touch-manipulation"  // 44px+ with icon

// ❌ Bad
className="p-1"  // Too small for touch
```

### 3. Truncation

Prevent text overflow on small screens:
```tsx
className="truncate max-w-[100px]"
```

### 4. Flexible Layouts

Use flex and min-width to prevent layout breaks:
```tsx
className="flex-1 min-w-0"  // Allows flex item to shrink
```

### 5. Conditional Rendering

Hide non-essential features on mobile:
```tsx
{/* Desktop only */}
<div className="hidden lg:block">
  <AdvancedFeature />
</div>
```

## Testing Responsive Design

### Browser DevTools

1. Open Chrome/Firefox DevTools
2. Toggle device toolbar (Cmd/Ctrl + Shift + M)
3. Test different device sizes:
   - iPhone SE (375px)
   - iPhone 12 Pro (390px)
   - iPad (768px)
   - Desktop (1024px+)

### Breakpoint Testing

Test at these critical widths:
- 375px (Small phone)
- 640px (sm breakpoint)
- 768px (md breakpoint)
- 1024px (lg breakpoint)
- 1280px (xl breakpoint)

### Touch Testing

On actual devices:
- Test tap targets
- Verify touch feedback
- Check scroll behavior
- Test overlay interactions

## Common Issues & Solutions

### Issue: Text Overflow

**Problem:** Long usernames or room names break layout

**Solution:**
```tsx
className="truncate max-w-[200px]"
```

### Issue: Small Touch Targets

**Problem:** Buttons too small on mobile

**Solution:**
```tsx
className="p-2 sm:p-3 touch-manipulation"
```

### Issue: Sidebar Not Closing

**Problem:** Mobile sidebar stays open after selection

**Solution:**
```tsx
onClick={() => {
  handleSelect();
  setIsMobileMenuOpen(false);
}}
```

### Issue: Layout Shift

**Problem:** Content jumps when sidebar opens

**Solution:**
```tsx
className="fixed lg:relative"  // Fixed on mobile, relative on desktop
```

## Future Enhancements

Potential improvements:
1. **Swipe Gestures:** Swipe to open/close sidebar
2. **Orientation Support:** Landscape mode optimizations
3. **Foldable Devices:** Support for foldable screens
4. **PWA Features:** Install prompt, offline support
5. **Accessibility:** Better screen reader support
6. **Performance:** Lazy load components on mobile

## Related Documentation

- [Loading Indicators](./LOADING_INDICATORS.md) - Loading states
- [Error Handling](./ERROR_HANDLING.md) - Error display
- [Optimistic Updates](./OPTIMISTIC_UPDATES.md) - Message sending
