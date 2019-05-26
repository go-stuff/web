# Changes to Bulma

- Removed `padding: 1.25rem;`
- added `padding: 10px;`

```css
.box {
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 2px 3px rgba(10, 10, 10, 0.1), 0 0 0 1px rgba(10, 10, 10, 0.1);
  color: #4a4a4a;
  display: block;
  /* padding: 1.25rem; */
  padding: 10px;
}
```

-Removed `margin-bottom: 1.5rem !important;`
-Added `margin-bottom: 0.0rem !important;`

```css
.tile.is-vertical > .tile.is-child:not(:last-child) {
  /* margin-bottom: 1.5rem !important; */
  margin-bottom: 0.0rem !important;
}
```