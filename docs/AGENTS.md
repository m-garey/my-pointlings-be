Fetch MyPointling project

User-customizable Pointling creatures that interact with the user and can be leveled up through scanning receipts and playing games.
This document outlines the MVP for the MyPointling project, which is a gamified feature within the Fetch app. The goal is to create a personalized experience for users by allowing them to interact with and customize their own Pointling creatures, earning points through various activities.

Motivations:

    avenue to spend points outside of Gift Cards

    way for users to personalize their experience

    get users to have a positive association w/ app

    vibes

MVP:

    Interactions (tamagotchi-esque things)

        feeding:

            earning points

        playing:

            via tab in fetch play (mischief)

        personality gets developed based on user interactions

    Customization (colors, outfits, naming)

        entry point

            naming, colors, common features [*]

            starting with neutral “Noot” pointling look

        leveling up

            access more accessories and features. features based on what you buy (AI)

                accessories = hats, shoes, face cover, limb colors(?) etc → access on level up, purchased with points

                features = colors, physical shape changes (wings, leaf, crab hands) → instantly obtain on level up, decided based on user interactions



Interactions

    XP to level up

    XP gain from Playing and Scanning, little bit of XP from checking on them in the app.

        Users shouldn’t have to do anything extra to get points. This is purely incentive

        Weights are dynamic: if you scan a buncha receipts, your next Play milestone is worth 5x or something

        Can only get benefits from 5 receipts a day

    FIDOs have categories (put in AI here tbh)

    Colors could be based off purchase type (like Apple Card)

    XP required to level up grows linearly from 3 XP, caps at 120 XP (to keep things fresh).

    Non-MVP potential for rebirths whatever to keep it fresh

    Gain XP: per receipt scan & per milestone reached

Customizations

    Features (colors, claws/dinos/wings)

    Accessories (hats, limb colors, shoes, capes, backgrounds + environmental objects, items)

    Accessories and features have various rarities

        Colors do not

    Obtaining features & accessories

        Level up

            3 random feature & accessory options, user picks 1

            Color every 5 levels (and level 2)

        Shop

            Directly buy some accessories

Tasks

    Backend

        Databases

            user: pointling, point amount

            pointling: current XP, required XP, level, current look, owned accessories

            accessories: item ID, asset ID, rarity, name, cost, type

            features: item ID, asset ID, rarity, name, cost, type

        API

            XP gain

            purchase

            2 level up endpoints (needs to return 3 items and validate 1 choice)

            various other things to hit database

            AI stuff lives here

    Frontend (Android)

        debug menu → trigger Receipt, Play events, add points

        “home” → customized Pointling + current level

        “wardrobe” → display all owned accessories and features

            lets you change them

        “shop” → spend Fetch Points on accessories

        “level up” → choose 1 of 3 items from backend
