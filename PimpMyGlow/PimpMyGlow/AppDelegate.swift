//
//  AppDelegate.swift
//  PimpMyGlow
//
//  Created by Mark Probst on 12/19/15.
//  Copyright © 2015 Mark Probst. All rights reserved.
//

import Foundation
import Cocoa

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, FileDragDestinationDelegate {
    @IBOutlet var window: NSWindow!
    @IBOutlet var gloDragDestination: FileDragDestination!
    @IBOutlet var aupDragDestination: FileDragDestination!
    @IBOutlet var destinationDragDestination: FileDragDestination!
    @IBOutlet var runButton: NSButton!
    @IBOutlet var clubsTextField: NSTextField!

    func applicationDidFinishLaunching(aNotification: NSNotification) {
        NSBundle.mainBundle().loadNibNamed("MainWindow", owner: self, topLevelObjects: nil)
        gloDragDestination.delegate = self
        aupDragDestination.delegate = self
        destinationDragDestination.delegate = self
        destinationDragDestination.mustBeDirectory = true
    }

    func draggedFile(sender: FileDragDestination, path: String) {
        runButton.enabled = gloDragDestination.path != nil && aupDragDestination.path != nil && destinationDragDestination.path != nil
    }

    func executeCommand(command: String, args: [String]) -> (Bool, String) {
        let task = NSTask()
        
        task.launchPath = command
        task.arguments = args
        
        let pipe = NSPipe()
        task.standardOutput = pipe
        task.standardError = pipe
        task.launch()
        task.waitUntilExit()
        
        let status = task.terminationStatus
        
        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        if let output = NSString(data: data, encoding: NSUTF8StringEncoding) as? String {
            return (status == 0, output)
        }
        
        return (status == 0, "")
    }

    @IBAction func runButtonPushed(sender: NSButtonCell) {
        let numClubs = Int(clubsTextField.stringValue)
        if numClubs == nil || numClubs! <= 0 {
            let alert = NSAlert()
            alert.messageText = "Number of clubs invalid"
            alert.informativeText = "Please enter a positive number"
            alert.addButtonWithTitle("OK")
            alert.runModal()
            return
        }
        
        let gloPath = gloDragDestination.path!
        let aupPath = aupDragDestination.path!
        let destinationPath = destinationDragDestination.path!
        Swift.print("run!", gloPath, aupPath, destinationPath)
        
        for club in 1...numClubs! {
            let executable = NSBundle.mainBundle().resourcePath! + "/glo-annotate"
            
            let (success, out) = executeCommand(executable,
                args: ["-audacity", aupPath, "-input", gloPath, "-club", "\(club)", "-output", destinationPath + "/\(club).glo"])
            if !success {
                let alert = NSAlert()
                alert.messageText = "Could not pimp for club \(club)"
                alert.informativeText = out
                alert.addButtonWithTitle("OK")
                alert.runModal()
                return
            }
        }
        
        let alert = NSAlert()
        alert.messageText = "Pimped!"
        alert.addButtonWithTitle("Yay!")
        alert.runModal()
        return
    }
}
