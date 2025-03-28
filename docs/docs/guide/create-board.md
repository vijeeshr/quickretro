# Create Board

The first thing you do is create/setup a board.\
Enter a name for the Board and an optional Team name.

::: info NOTE
The board creator is also the board owner and can perform multiple actions not available to others.\
We'll soon see it in [Dashboard](dashboard) section.
:::

## Configuring Board Columns
::: tip
Since the introduction of multi-language support with <Badge type="tip" text="v1.3.0" />, default column names can be automatically translated to other
languages.\
***Custom column names are not automatically translated.***\
It is recommended to use the defaults, if any of your team members use the app in a different language.
:::

A max of 5 columns are allowed. The first 3 columns are always enabled by default.\
You can choose which columns you want and name them accordingly.

<img src="/createboard.png" class="shadow-img" alt="Create Board" width="360" loading="lazy">

Click the coloured dot (***present towards left of each column name***) to enable/disable a column.\
Click the column name text to type any custom name.

When a Board is created, the user is taken to the [Dashboard](dashboard).

## Quick video

<video class="video-play" id="createBoardVideo" controls width="640">
  <source src="/videos/create-board.mp4" type="video/webm">
  Your browser does not support the video tag.
</video>

## Cloudflare Turnstile Integration

Available from <Badge type="tip" text="v1.4.0" />

<img src="/createboard_turnstile.png" class="shadow-img" alt="Cloudflare Turnstile" width="360" loading="lazy">

Cloudflare Turnstile is a CAPTCHA alternative provided by Cloudflare. The integration can be enabled/disabled in a configurable way. It is disabled by default. 

Details to enable it provided in [Configurations](configurations#enable-cloudflare-turnstile)

<script setup>
import { onMounted } from 'vue';

onMounted(() => {
  const video = document.getElementById('createBoardVideo');
  if (video) {
    video.playbackRate = 2.5; // Adjust speed (e.g., 1.5x, 2x)
  }
});
</script>