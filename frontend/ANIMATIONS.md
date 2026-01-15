# Animations and Transitions

This document describes the animations and transitions implemented throughout the chat application to enhance user experience and provide visual feedback.

## Overview

The application uses a combination of CSS animations and transitions to create smooth, polished interactions. All animations are performance-optimized and respect user preferences for reduced motion.

## Animation System

### Tailwind Custom Animations

Custom animations are defined in `tailwind.config.js`:

```javascript
animation: {
  'fadeIn': 'fadeIn 0.3s ease-in',
  'slideInLeft': 'slideInLeft 0.3s ease-out',
  'slideInRight': 'slideInRight 0.3s ease-out',
  'slideDown': 'slideDown 0.2s ease-out',
  'slideUp': 'slideUp 0.2s ease-out',
  'scaleIn': 'scaleIn 0.2s ease-out',
  'bounce-subtle': 'bounce-subtle 1s ease-in-out infinite',
}
```

### Keyframe Definitions

**fadeIn**: Smooth opacity transition
```css
0%: opacity 0
100%: opacity 1
```

**slideInLeft**: Slide from left with fade
```css
0%: translateX(-20px), opacity 0
100%: translateX(0), opacity 1
```

**slideInRight**: Slide from right with fade
```css
0%: translateX(20px), opacity 0
100%: translateX(0), opacity 1
```

**slideDown**: Slide down with fade
```css
0%: translateY(-10px), opacity 0
100%: translateY(0), opacity 1
```

**slideUp**: Slide up with fade
```css
0%: translateY(10px), opacity 0
100%: translateY(0), opacity 1
```

**scaleIn**: Scale up with fade
```css
0%: scale(0.95), opacity 0
100%: scale(1), opacity 1
```

**bounce-subtle**: Gentle bounce effect
```css
0%, 100%: translateY(0)
50%: translateY(-3px)
```

## Component Animations

### Message Animations

**Location:** `MessageItem.tsx`

**Animations:**
- Fade in on mount
- Slide in from left (received messages)
- Slide in from right (sent messages)

**Implementation:**
```tsx
className="animate-fadeIn"  // Container
className="animate-slideInLeft"  // Received messages
className="animate-slideInRight"  // Sent messages
```

**Purpose:**
- Provides visual feedback when new messages arrive
- Distinguishes between sent and received messages
- Creates smooth, natural message flow

### Typing Indicator Animation

**Location:** `TypingIndicator.tsx`

**Animations:**
- Slide down on appear
- Bouncing dots (staggered)

**Implementation:**
```tsx
className="animate-slideDown"  // Container
className="animate-bounce"  // Dots with delay
style={{ animationDelay: '0ms' }}  // First dot
style={{ animationDelay: '150ms' }}  // Second dot
style={{ animationDelay: '300ms' }}  // Third dot
```

**Purpose:**
- Indicates active typing
- Creates engaging visual feedback
- Staggered animation feels natural

### Modal Animations

**Location:** `CreateRoomModal.tsx`, `RoomSettings.tsx`, `NotificationList.tsx`

**Animations:**
- Fade in overlay
- Scale in modal content

**Implementation:**
```tsx
className="animate-fadeIn"  // Overlay
className="animate-scaleIn"  // Modal content
```

**Purpose:**
- Smooth modal appearance
- Draws attention to modal
- Professional feel

### Room List Item Animations

**Location:** `RoomList.tsx`

**Animations:**
- Hover scale effect
- Shadow transition
- Background color transition

**Implementation:**
```tsx
className="transition-all duration-200 ease-in-out transform hover:scale-[1.02]"
```

**Purpose:**
- Provides hover feedback
- Indicates interactivity
- Smooth selection experience

### Button Animations

**Locations:** Throughout the app

**Animations:**
- Scale on hover
- Active state scale down
- Color transitions
- Special effects (rotate for create button)

**Implementation:**
```tsx
// Standard button
className="transition-all duration-200 ease-in-out transform hover:scale-110 active:scale-95"

// Create button (with rotation)
className="transform hover:scale-110 hover:rotate-90"

// Send button
className="transform hover:scale-105 active:scale-95"
```

**Purpose:**
- Provides tactile feedback
- Indicates clickability
- Enhances user engagement

### Sidebar Animations

**Location:** `ChatPage.tsx`

**Animations:**
- Slide in from left (mobile)
- Fade in overlay

**Implementation:**
```tsx
className="animate-slideInLeft"  // Sidebar
className="animate-fadeIn"  // Overlay
```

**Purpose:**
- Smooth mobile menu appearance
- Clear visual hierarchy
- Natural interaction flow

## Transition Classes

### Standard Transitions

**Color Transitions:**
```tsx
className="transition-colors duration-200"
```
- Used for: Background colors, text colors, borders
- Duration: 200ms
- Easing: Default (ease)

**All Properties:**
```tsx
className="transition-all duration-200 ease-in-out"
```
- Used for: Complex animations (scale + color + shadow)
- Duration: 200ms
- Easing: ease-in-out

**Transform Only:**
```tsx
className="transition-transform duration-200"
```
- Used for: Scale, rotate, translate
- Duration: 200ms
- Easing: Default (ease)

### Hover Effects

**Scale Up:**
```tsx
className="transform hover:scale-110"
```
- Used for: Buttons, icons
- Scale: 1.1x (10% larger)

**Scale Down (Active):**
```tsx
className="active:scale-95"
```
- Used for: Button press feedback
- Scale: 0.95x (5% smaller)

**Subtle Scale:**
```tsx
className="hover:scale-[1.02]"
```
- Used for: List items, cards
- Scale: 1.02x (2% larger)

**Rotate:**
```tsx
className="hover:rotate-90"
```
- Used for: Create button (plus icon)
- Rotation: 90 degrees

## Smooth Scrolling

**Location:** `MessageList.tsx`

**Implementation:**
```tsx
className="scroll-smooth"  // Container
scrollIntoView({ behavior: 'smooth', block: 'end' })  // JavaScript
```

**Purpose:**
- Smooth auto-scroll to new messages
- Better user experience
- Reduces jarring jumps

## Performance Considerations

### GPU Acceleration

Animations use transform and opacity for GPU acceleration:
```tsx
// ✅ Good (GPU accelerated)
transform: translateX(), scale(), rotate()
opacity

// ❌ Avoid (CPU intensive)
left, top, width, height
```

### Animation Duration

Optimal durations for different effects:
- **Quick feedback**: 100-200ms (buttons, hovers)
- **Standard animations**: 200-300ms (slides, fades)
- **Slow animations**: 300-500ms (complex transitions)

### Reduced Motion

Respect user preferences:
```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

## Animation Best Practices

### 1. Use Appropriate Timing

**Fast (100-200ms):**
- Button hovers
- Color changes
- Small scale changes

**Medium (200-300ms):**
- Slides
- Fades
- Modal appearances

**Slow (300-500ms):**
- Complex transitions
- Page transitions
- Large movements

### 2. Choose Right Easing

**ease-in:** Starts slow, ends fast
- Use for: Elements leaving the screen

**ease-out:** Starts fast, ends slow
- Use for: Elements entering the screen

**ease-in-out:** Slow start and end
- Use for: Elements moving within screen

**linear:** Constant speed
- Use for: Loading spinners, continuous animations

### 3. Combine Animations

Layer multiple effects for rich interactions:
```tsx
className="transition-all duration-200 transform hover:scale-110 hover:shadow-lg"
```

### 4. Provide Feedback

Always animate user interactions:
- Hover states
- Active/pressed states
- Loading states
- Success/error states

### 5. Keep It Subtle

Avoid excessive animation:
- Small scale changes (1.02x - 1.1x)
- Short durations (200-300ms)
- Gentle easing curves

## Common Animation Patterns

### Fade In on Mount

```tsx
<div className="animate-fadeIn">
  {content}
</div>
```

### Slide In from Direction

```tsx
// From left
<div className="animate-slideInLeft">
  {content}
</div>

// From right
<div className="animate-slideInRight">
  {content}
</div>
```

### Scale on Hover

```tsx
<button className="transition-transform duration-200 transform hover:scale-110">
  {label}
</button>
```

### Smooth Color Transition

```tsx
<div className="transition-colors duration-200 hover:bg-blue-500">
  {content}
</div>
```

### Loading Spinner

```tsx
<svg className="animate-spin">
  {/* spinner icon */}
</svg>
```

### Bounce Effect

```tsx
<div className="animate-bounce">
  {content}
</div>
```

## Testing Animations

### Visual Testing

1. **Check timing**: Animations should feel natural, not too fast or slow
2. **Check smoothness**: No jank or stuttering
3. **Check consistency**: Similar elements should animate similarly
4. **Check accessibility**: Respect reduced motion preferences

### Performance Testing

1. **Monitor FPS**: Should maintain 60fps
2. **Check CPU usage**: Animations shouldn't spike CPU
3. **Test on mobile**: Ensure smooth on lower-end devices
4. **Test with many elements**: Performance with 100+ messages

### Browser Testing

Test animations in:
- Chrome/Edge (Chromium)
- Firefox
- Safari
- Mobile browsers

## Troubleshooting

### Animation Not Working

**Check:**
1. Class name is correct
2. Tailwind config includes animation
3. CSS is rebuilt after config changes
4. No conflicting styles

### Animation Stuttering

**Solutions:**
1. Use transform instead of position
2. Use opacity instead of visibility
3. Add `will-change` for complex animations
4. Reduce animation complexity

### Animation Too Fast/Slow

**Adjust:**
1. Change duration in Tailwind config
2. Modify easing function
3. Test on different devices

## Future Enhancements

Potential improvements:
1. **Page Transitions**: Animate between routes
2. **Gesture Animations**: Swipe to delete, pull to refresh
3. **Micro-interactions**: More subtle feedback animations
4. **Loading Skeletons**: Animated placeholders
5. **Confetti Effects**: Celebration animations
6. **Parallax Effects**: Depth and dimension

## Related Documentation

- [Responsive Design](./RESPONSIVE_DESIGN.md) - Layout adaptations
- [Loading Indicators](./LOADING_INDICATORS.md) - Loading states
- [Optimistic Updates](./OPTIMISTIC_UPDATES.md) - Message animations
