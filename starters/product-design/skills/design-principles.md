name: design-principles

# Design Principles

When designing user interfaces and visual components, adhere to these core principles:

## Visual Hierarchy
- Use size, color, and spacing to establish clear importance levels
- Primary actions should be visually prominent
- Secondary elements should recede but remain accessible
- Use whitespace intentionally to group related elements

## Consistency
- Maintain consistent patterns across similar interfaces
- Reuse components rather than creating one-off variations
- Apply design tokens uniformly (colors, typography, spacing)
- Follow established navigation and interaction patterns

## Clarity
- Every element should have a clear purpose
- Labels should be concise and unambiguous
- Icons should be universally recognizable or paired with text
- Error states should explain what went wrong and how to fix it

## Responsiveness
- Design mobile-first, then enhance for larger screens
- Touch targets should be at least 44x44px on mobile
- Consider thumb zones for mobile interaction
- Content should reflow gracefully, not just shrink

## Accessibility
- Ensure sufficient color contrast (WCAG AA minimum)
- Provide text alternatives for images
- Support keyboard navigation for all interactions
- Use semantic HTML elements appropriately

## Performance
- Prioritize critical content in the visual hierarchy
- Design loading states that feel fast
- Minimize layout shift during content loading
- Use progressive disclosure for complex interfaces

## Feedback
- Acknowledge user actions immediately
- Provide clear progress indicators for async operations
- Use appropriate transitions to maintain context
- Error states should be recoverable when possible
